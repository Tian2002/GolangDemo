package reply

//gedis/resp/reply/reply.go

import (
	"bytes"
	"gedis/interface/resp"
	"strconv"
)

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

var (
	nullBulkReplyBytes = []byte("$-1\r\n")
	CRLF               = "\r\n"
)

// BulkReply 返回多行字符串，以“$”开头，中间跟字节数,“\r\n”结尾
// 发送“xzh“，即“$3\r\nxzh\r\n”
type BulkReply struct {
	Arg []byte
}

func (r *BulkReply) ToBytes() []byte {
	l := len(r.Arg)
	if l == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(l) + CRLF + string(r.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// MultiBulkReply 返回数组，以“*”开头，后面跟成员的个数，“\r\n”结尾
// 发送“set key value“，即“*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
type MultiBulkReply struct {
	Args [][]byte
}

func (r *MultiBulkReply) ToBytes() []byte {
	l := len(r.Args)
	var buf bytes.Buffer //这样写效率更高
	buf.WriteString("*" + strconv.Itoa(l) + CRLF)
	for i := 0; i < l; i++ {
		buf.WriteString(string((&BulkReply{Arg: r.Args[i]}).ToBytes()))
	}
	return buf.Bytes()
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

// StatusReply 返回一些简单的状态
type StatusReply struct {
	Status string
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

// IntReply 以“:”开头，“\r\n”结尾
type IntReply struct {
	Code int64
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

// StandardErrReply 返回一些标准的错误
type StandardErrReply struct {
	Status string
}

func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// IsErrorReply 判断这个reply是不是错误
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
