package cluster

import (
	"context"
	"errors"
	"gedis/interface/resp"
	"gedis/resp/client"
)

//gedis/cluster/com.go
import (
	"gedis/lib/utils"
	"gedis/resp/reply"
	"strconv"
)

//负责和节点间的通信

// 从连接池拿到一个连接
func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New("没有这个连接")
	}
	//借用一个连接
	object, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	c, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("连接池中保持的连接类型错误")
	}
	return c, err
}

// 把连接池的连接还回去
func (cluster *ClusterDatabase) returnPeerClient(peer string, c *client.Client) error {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("没有这个连接")
	}
	//归还
	return pool.ReturnObject(context.Background(), c)
}

// 转发请求
func (cluster *ClusterDatabase) relay(peer string, conn resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self { //如果是我们自己
		return cluster.db.Exec(conn, args)
	}
	cli, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}
	defer cluster.returnPeerClient(peer, cli)
	//每次执行时要先切换数据库
	cli.Send(utils.ToCmdLine("select", strconv.Itoa(conn.GetDBIndex())))

	return cli.Send(args)
}

// 广播请求
func (cluster *ClusterDatabase) broadcast(conn resp.Connection, args [][]byte) map[string]resp.Reply {
	m := make(map[string]resp.Reply)
	for _, peer := range cluster.nodes {
		relay := cluster.relay(peer, conn, args)
		m[peer] = relay
	}
	return m
}
