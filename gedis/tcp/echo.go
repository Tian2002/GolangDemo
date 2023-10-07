package tcp

import (
	"bufio"
	"context"
	"gedis/lib/logger"
	"gedis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (e *EchoClient) Close() error {
	e.Waiting.WaitWithTimeout(10 * time.Second)
	return e.Conn.Close()
}

// EchoHandler 一个简单的业务，用户向我们发送什么，我们就回复什么
type EchoHandler struct {
	activeConn sync.Map //记录有多少连接,这里是当set使用
	closing    atomic.Bool
}

func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if handler.closing.Load() { //当业务引擎在关机的过程中，我们直接将要连接的客户端关闭
		conn.Close()
	}
	//先把连接包装为我们内部的一个结构体
	client := &EchoClient{
		Conn: conn,
	}
	//将client记录下来
	handler.activeConn.Store(client, struct{}{})
	//服务客户
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n') //规定换行时一个命令的结束
		if err != nil {
			if err == io.EOF { //数据结束符
				logger.Info("echo客服端连接关闭")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1) //我们正在做业务，不要关掉我，
		conn.Write([]byte(msg))
		client.Waiting.Done() //业务完成，可以关掉
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("业务正在关闭")
	handler.closing.Store(true)
	//业务引擎要关闭了，将所有记录的客户端关掉
	handler.activeConn.Range(func(key, value any) bool {
		client := key.(*EchoClient)
		client.Conn.Close()
		return true
	})
	return nil
}
