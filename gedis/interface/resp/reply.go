package resp

// Reply 代表各种对客服端的回复
type Reply interface {
	ToBytes() []byte
}
