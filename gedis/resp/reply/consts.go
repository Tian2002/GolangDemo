package reply

//gedis/resp/reply/consts.go
//固定的回复

type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")
var thePongReply = new(PongReply) //在本地持有一个该类型的指针，避免每次都创建一个该类型的指针，节约内存
func (r PongReply) ToBytes() []byte {
	return pongBytes
}
func MakePongReply() *PongReply {
	return thePongReply
}

type OKReply struct{}

var OKBytes = []byte("+OK\r\n")
var theOkReply = new(OKReply)

func (r *OKReply) ToBytes() []byte {
	return OKBytes
}
func MakeOKReply() *OKReply {
	return theOkReply
}

// NullBulkReply 空的字符串回复,是nil，而不是""
type NullBulkReply struct {
}

var nullBulkBytes = []byte("$-1\r\n")
var theNullBulkReply = new(NullBulkReply)

func (r NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}
func MakeNullBulkReply() *NullBulkReply {
	return theNullBulkReply
}

// EmptyMultiBulkReply 这里是返回空数组
type EmptyMultiBulkReply struct {
}

var emptyMultiBulkBytes = []byte("*0\r\n")
var theEmptyMultiBulkReply = new(EmptyMultiBulkReply)

func (r EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}
func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return theEmptyMultiBulkReply
}

// NoReply 回复一个真的空
type NoReply struct{}

var noReplyBytes = []byte("")
var theNoReply = new(NoReply)

func (r NoReply) ToBytes() []byte {
	return noReplyBytes
}
func MakeNoReply() *NoReply {
	return theNoReply
}
