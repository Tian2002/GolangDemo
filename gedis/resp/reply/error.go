package reply

//gedis/resp/reply/error.go

// UnKnowErrReply 未知错误
type UnKnowErrReply struct{}

var unKnowErrBytes = []byte("-Err unKnow\r\n")
var theUnKnowErrReply = new(UnKnowErrReply)

func (r UnKnowErrReply) Error() string {
	return string(unKnowErrBytes)
}

func (r UnKnowErrReply) ToBytes() []byte {
	return unKnowErrBytes
}

func MakeUnKnowErrReply() *UnKnowErrReply {
	return theUnKnowErrReply
}

// ArgNumErrReply 客户端发送到置零参数个数错误
type ArgNumErrReply struct {
	Cmd string //记录一下指令
}

func (r *ArgNumErrReply) Error() string {
	return string(r.ToBytes())
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-Err wrong number of arguments for'" + r.Cmd + "' command\r\n")
}
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{Cmd: cmd}
}

// SyntaxErrReply 语法错误
type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-Err syntax error\r\n")
var theSyntaxErrReply = new(SyntaxErrReply)

func (r SyntaxErrReply) Error() string {
	return string(syntaxErrBytes)
}

func (r SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

// WrongTypeErrReply 数据类型错误
type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-Err WrongType Operation against a key holding the wrong kind of value\r\n")
var theWrongTypeErrReply = new(WrongTypeErrReply)

func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return string(r.ToBytes())
}
func MakeWrongTypeErrReply() *WrongTypeErrReply {
	return theWrongTypeErrReply
}

// ProtocolErrReply 协议错误
type ProtocolErrReply struct {
	Msg string
}

func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n")
}

func (r *ProtocolErrReply) Error() string {
	return string(r.ToBytes())
}
func MakeProtocolErrReply(msg string) *ProtocolErrReply {
	return &ProtocolErrReply{Msg: msg}
}
