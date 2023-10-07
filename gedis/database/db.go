package database

//gedis/database/db.go
import (
	"gedis/datastruct/dict"
	"gedis/interface/database"
	"gedis/interface/resp"
	"gedis/resp/reply"
	"strings"
)

type DB struct {
	index  int
	data   dict.Dict
	addAof func(line CmdLine)
}

// ExecFunc 方法执行函数
type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine = [][]byte

func makeDB() *DB {
	return &DB{data: dict.MakeSyncDict(), addAof: func(line CmdLine) {
	}}
}

func (db *DB) Exec(conn resp.Connection, cmdLine CmdLine) resp.Reply {
	//查看是哪个命令，如PING SET SETNX
	cmdName := strings.ToLower(string(cmdLine[0]))
	command, exists := cmdTable[cmdName]
	if !exists { //命令错误或者命令未注册
		return reply.MakeErrReply("Err UnKnown reply" + cmdName)
	}
	if !validateArity(command.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	//这里切一下是因为我们已经找到了是哪个指令，只需要传递参数就可以了 set k v -> k v
	return command.exector(db, cmdLine[1:])
}

// 校验参数个数是否合法
func validateArity(arity int, cmdArgs [][]byte) bool {
	//TODO valid
	return true
}

//下面是实现一些数据存取的操作封装，我们直接操作DB就可以实现简单的存取，而不需要去操作DB的成员

// GetEntity 获取值
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {

	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

// PutEntity 把键值对放入数据库
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists 存在才放入
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

// PutIfAbsent 不存在才放入
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove 从数据库移除单个数据
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// Removes 从数据库移除多个数据
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

// Flush 删除数据库内的所有内容
func (db *DB) Flush() {
	db.data.Clear()
}
