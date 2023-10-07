package cluster

//gedis/cluster/router.go
import "gedis/interface/resp"

// 默认的转发方法
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply {
	//拿到一致性哈希的值，通常是根据key
	key := string(cmdArgs[1])
	return cluster.relay(cluster.peerPicker.PickNode(key), conn, cmdArgs)
}

func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["ping"] = ping

	routerMap["del"] = Del

	routerMap["exists"] = defaultFunc
	routerMap["type"] = defaultFunc
	routerMap["rename"] = Rename
	routerMap["renamenx"] = Rename

	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getset"] = defaultFunc

	routerMap["flushdb"] = FlushDB

	return routerMap
}

var router = makeRouter()
