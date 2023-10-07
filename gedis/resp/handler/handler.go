package handler

/*
 * A tcp.RespHandler implements redis protocol
 */
//gedis/resp/handler/handler.go
import (
	"context"
	"gedis/cluster"
	"gedis/config"
	"gedis/database"
	databaseface "gedis/interface/database"
	"gedis/lib/logger"
	"gedis/resp/connection"
	"gedis/resp/parser"
	"gedis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

// RespHandler 实现接口tcp.Handler作为redis的handler
type RespHandler struct {
	activeConn sync.Map // *client -> placeholder
	db         databaseface.Database
	closing    atomic.Bool // refusing new client and new request
}

func MakeHandler() *RespHandler {
	var db databaseface.Database
	//db = database.NewStandaloneDatabase()
	//判断是否打开集群
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		db = cluster.MakeClusterDatabase()
	} else {
		db = database.NewStandaloneDatabase()
	}
	return &RespHandler{
		db: db,
	}
}

func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Handle 接收并执行redis命令，这里这个业务实现的很简单，就是打印发来的命令
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Load() {
		// closing handler refuse new connection
		_ = conn.Close()
	}

	client := connection.NewConn(conn)
	h.activeConn.Store(client, 1)

	ch := parser.ParseStream(conn)
	for payload := range ch { //把解析的指令输出
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok { //一般客服端的数据都是通过数组发来的，但是其他的我们也能解析，但是这里打印一个错误日志
			logger.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args) //这里是执行指令，操作数据库
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close stops handler
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Store(true)
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}
