package parser

//gedis/resp/parser/parser.go
import (
	"bufio"
	"errors"
	"gedis/interface/resp"
	"gedis/lib/logger"
	"gedis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

// Payload 客服端发送来的数据
type Payload struct {
	Data resp.Reply //这是使用服务端定义的结构是因为互相发送的消息的结构是一样的
	Err  error
}

// 解析器的状态
type readState struct {
	readingMultiLine  bool     //在解析单行还是多行数据，如果是多行则还没读取完成，如遇到"*"或"$"
	expectedArgsCount int      //正在读取指令应该有几个参数，字符串数组用的
	msgType           byte     //指令的类型
	args              [][]byte //传过来的解析后的数据
	bulkLen           int64    //解析的字符串的长度
}

// ParseStream 异步的去解析协议，把解析的结构放进一个channel中
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch) //将协议解析去异步执行
	return ch
}

// 计算解析有没有完成
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for { //循环解析，解析每一行
		// read line
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr { // 遇到io错误，停止读取，把已经读到的数据写入channel
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			//协议错误，重置读取的状态
			ch <- &Payload{
				Err: err,
			}
			state = readState{}
			continue
		}

		// 解析这一条的命令
		if !state.readingMultiLine { // 刚开始时，接收新响应或者state.readingMultiLine已经被标记为true
			if msg[0] == '*' { //依次判断，看看是RESP协议中五种数据的哪一种，这里是判断是不是数组
				// 解析multiBulk
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{} // reset state
					continue
				}
			} else if msg[0] == '$' { // 解析多行字符串
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.bulkLen == -1 { // null bulk reply
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{} // reset state
					continue
				}
			} else {
				// 解析正常回复、错误回复、整数回复
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{} // reset state
				continue
			}
		} else {
			// 收到以下批量回复
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{} // reset state
				continue
			}
			// if sending finished
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}
}

// 这个方法是读取一条的命令，第二个参数是标识当error不为空时，错误是否为io的错误
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 { // 只有解析到"$"符号后，state.bulkLen才会变，当正在解析"$"时也还是0
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { // read bulk line (binary safe)
		msg = make([]byte, state.bulkLen+2)  //state.bulkLen只标识了字符串和的长度，所有还要加上"\r\n"的长度
		_, err = io.ReadFull(bufReader, msg) //这里使用ReadFull, 而不继续使用 ReadBytes('\n') 读取下一行, 这是因为在字符串内部可能包含"\r\n",这里应该特殊处理
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0 //解析完成后重置state.bulkLen，因为下一行又是$开头或者直接结束
	}
	return msg, false, nil
}

func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// first line of multi bulk reply
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+': // status reply
		result = reply.MakeStatusReply(str[1:])
	case '-': // err reply
		result = reply.MakeErrReply(str[1:])
	case ':': // int reply
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// 阅读多批量回复或批量回复的非第一行，即1、读到了"*"开头的后面 或 2、先前读到了"$"开头的后面
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' { //对应1
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else { //对应2
		state.args = append(state.args, line)
	}
	return nil
}
