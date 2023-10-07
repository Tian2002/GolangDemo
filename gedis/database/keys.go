package database

//gedis/database/keys.go
import (
	"gedis/interface/resp"
	"gedis/lib/utils"
	"gedis/lib/wildcard"
	"gedis/resp/reply"
)

// DEL 删除数据，可以一次删除多个
func execDel(db *DB, args [][]byte) resp.Reply {
	var count int64
	for _, v := range args { //这里不删掉第一个，这是因为参数类型已经在上层去掉了
		db.Remove(string(v))
		count++
	}
	if count > 0 { //这里要加上命令参数
		db.addAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(count)
}

// EXISTS
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// TYPE
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
		//TODO 其他类型
	}
	return reply.MakeUnKnowErrReply()
}

// RENAME
func execRename(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[0])
	dest := string(args[1])

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine2("rename", args...))
	return reply.MakeOKReply()
}

// RENAMENX rename key1 key2 重命名如果不存在key2, 避免原本有key2，将key2覆盖
func execRenameNx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	db.Removes(src, dest) // clean src and dest with their ttl
	db.PutEntity(dest, entity)
	db.addAof(utils.ToCmdLine2("renamenx", args...))
	return reply.MakeIntReply(1)
}

// FLUSH
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLine2("flushdb", args...))
	return reply.MakeOKReply()
}

// KEYS
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	//-2表示最少是两个参数，如del key1，但是是变长的，这里我们就用负数来表示
	RegisterCommand("Del", execDel, -2)
	RegisterCommand("Exists", execExists, -2)
	RegisterCommand("FlushDB", execFlushDB, -1)
	RegisterCommand("Type", execType, 2)
	RegisterCommand("Rename", execRename, 3)
	RegisterCommand("RenameNx", execRenameNx, 3)
	RegisterCommand("Keys", execKeys, 2)
}
