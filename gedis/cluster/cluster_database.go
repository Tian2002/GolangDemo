package cluster

//gedis/cluster/cluster_database.go
import (
	"context"
	"gedis/config"
	database2 "gedis/database"
	"gedis/interface/database"
	"gedis/interface/resp"
	"gedis/lib/consistenthash"
	"gedis/lib/logger"
	"gedis/resp/reply"
	pool "github.com/jolestar/go-commons-pool/v2"
	"strings"
)

type ClusterDatabase struct {
	self           string //自己的地址
	nodes          []string
	peerPicker     *consistenthash.NodeMap     //节点选择器
	peerConnection map[string]*pool.ObjectPool //保存多个连接池，对每个兄弟节点都有一个连接
	db             database.Database
}

func MakeClusterDatabase() *ClusterDatabase {
	clister := &ClusterDatabase{
		self:           config.Properties.Self,
		db:             database2.NewStandaloneDatabase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}

	nodes := config.Properties.Peers
	nodes = append(nodes, config.Properties.Self)
	clister.nodes = nodes
	clister.peerPicker.AddNode(nodes...)
	//对一个个兄弟节点新建连接池
	for _, peer := range config.Properties.Peers {
		objectPool := pool.NewObjectPoolWithDefaultConfig(context.Background(), &connectionFactory{Peer: peer})

		clister.peerConnection[peer] = objectPool
	}

	return clister
}

func (c *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	defer func() { //防止核心业务因为未知业务而停止
		err := recover()
		if err != nil {
			logger.Error(err)
			result = reply.MakeUnKnowErrReply()
		}
	}()
	cmdName := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		result = reply.MakeErrReply("错误的命令或暂时未实现的命令")
	}
	result = cmdFunc(c, client, args)
	return result
}

// AfterClientClose 其实执行单机版的方法
func (c *ClusterDatabase) AfterClientClose(conn resp.Connection) {
	c.db.AfterClientClose(conn)
}

// Close 其实关掉单机版的db
func (c *ClusterDatabase) Close() {
	c.db.Close()
}

type CmdFunc func(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply
