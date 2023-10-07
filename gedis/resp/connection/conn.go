package connection

//gedis/resp/connection/conn.go
import (
	"gedis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

// Connection 表示使用redis cli的连接
type Connection struct {
	conn net.Conn
	// 等待答复完成
	waitingReply wait.Wait
	// 在程序发送响应时锁定
	mu sync.Mutex
	// 选择数据库
	selectedDB int
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// RemoteAddr 返回远程网络地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close 断开与客户端的连接
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// Write 通过tcp连接向客户端发送响应，把数据写回去
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()

	_, err := c.conn.Write(b)
	return err
}

// GetDBIndex 返回现在在使用的数据库编号
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB 通过编号选择一个数据库
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
