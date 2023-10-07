package cluster

//gedis/cluster/rename.go
import (
	"gedis/interface/resp"
	"gedis/resp/reply"
)

// Rename 重命名键，源和目标必须在同一节点内,真的要实现也比较简单，就是重新包装这个命令，在把src里面的删除，在desc中添加
func Rename(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[1])
	dest := string(args[2])

	srcPeer := cluster.peerPicker.PickNode(src)
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within one slot in cluster mode")
	}
	return cluster.relay(srcPeer, c, args)
}
