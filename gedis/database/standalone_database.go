package database

//gedis/database/standalone_database.go
import (
	"fmt"
	"gedis/aof"
	"gedis/config"
	"gedis/interface/resp"
	"gedis/lib/logger"
	"gedis/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

// StandaloneDatabase 在Redis中默认是有16个db的，这里就是将多个数据库记录下来
type StandaloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

// NewStandaloneDatabase 创建每个数据库
func NewStandaloneDatabase() *StandaloneDatabase {
	mdb := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	mdb.dbSet = make([]*DB, config.Properties.Databases)
	for i := range mdb.dbSet {
		singleDB := makeDB()
		singleDB.index = i
		mdb.dbSet[i] = singleDB
	}
	//aof初始化
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(mdb)
		if err != nil {
			panic(err)
		}
		mdb.aofHandler = aofHandler
		//将每个db的函数成员初始化
		for _, db := range mdb.dbSet {
			//这里是一个闭包，会逃逸到堆上，导致所有的dbIndex都是15,当然这也是for range的坑
			singleDB := db
			singleDB.addAof = func(line CmdLine) {
				mdb.aofHandler.AddAof(singleDB.index, line)
			}
		}
	}
	return mdb
}

// Exec executes command
// 参数“cmdLine”包含命令及其参数，例如：“set key value”
func (mdb *StandaloneDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil { //这是整个数据库的核心逻辑的上层，避免数据库在操作中出现未知的错误，抛出panic，导致程序崩溃，这里recover
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "select" { //select是其中特色的一个命令，是切换数据库，而不是操作某一个数据库
		if len(cmdLine) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(c, mdb, cmdLine[1:])
	}
	// 一般的命令，直接调用 db.xxx
	dbIndex := c.GetDBIndex()
	selectedDB := mdb.dbSet[dbIndex]
	return selectedDB.Exec(c, cmdLine)
}

// Close 这个和下面那个都没有特殊的处理
func (mdb *StandaloneDatabase) Close() {
}

func (mdb *StandaloneDatabase) AfterClientClose(c resp.Connection) {
}

// 执行选择数据库的逻辑
func execSelect(c resp.Connection, mdb *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(mdb.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOKReply()
}
