package database

//gedis/database/ping.go
import (
	"gedis/interface/resp"
	"gedis/resp/reply"
)

// Ping 实现数据库内的操作以及相对应的返回
func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// 还需要在程序启动时注册Ping指令
func init() {
	RegisterCommand("ping", Ping, 1)
}
