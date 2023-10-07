package aof

//gedis/aof/aof.go
import (
	"gedis/config"
	"gedis/interface/database"
	"gedis/lib/logger"
	"gedis/lib/utils"
	"gedis/resp/connection"
	"gedis/resp/parser"
	"gedis/resp/reply"
	"io"
	"os"
	"strconv"
)

const aofBufSize = 1 << 16

// AofHandler aof文件处理器
type AofHandler struct {
	database    database.Database
	aofFile     *os.File
	aofFileName string
	//当前写入的DB,这个字段的目的是为了不必再每条指令前都加上select语句，而是在切换数据库时才加上
	currentDB int
	//aof写文件的缓冲区，异步的写文件
	aofChan chan *payload
}

type payload struct {
	commandLine [][]byte
	dbIndex     int
}

// NewAofHandler 新建一个aofHandler
func NewAofHandler(database database.Database) (*AofHandler, error) {
	handler := new(AofHandler)
	handler.aofFileName = config.Properties.AppendFilename
	handler.database = database
	//加载aof文件
	handler.LoadAof()

	aofFile, err := os.OpenFile(handler.aofFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	//defer aofFile.Close()
	handler.aofFile = aofFile
	//初始化channel
	handler.aofChan = make(chan *payload, aofBufSize)
	//异步的写入文件
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// AddAof 把追加的内容放入缓冲区,不同步落盘，异步操作
func (handler *AofHandler) AddAof(dbIndex int, cmd [][]byte) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			commandLine: cmd,
			dbIndex:     dbIndex,
		}
	}
}

// 把缓冲区的内容追加到硬盘里
func (handler *AofHandler) handleAof() {
	//在启动时将默认数据库设置为0，避免上次关机前的aof文件中最后使用的不是0号数据库
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB { //切换了DB，需要加上select指令
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Error(err)
				continue
			}
			p.dbIndex = handler.currentDB
		}
		//将真正的指令落盘
		data := reply.MakeMultiBulkReply(p.commandLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Error(err)
			continue
		}
	}
}

// LoadAof 加载文件，在初始化重启的时候执行
func (handler *AofHandler) LoadAof() {
	aofFile, err := os.Open(handler.aofFileName)
	if err != nil {
		logger.Error(err)
		return
	}
	defer aofFile.Close()
	ch := parser.ParseStream(aofFile)
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			} else {
				logger.Error(err)
				continue
			}
		}
		if p.Data == nil {
			logger.Error("payload empty err")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("payload Data err,need multi bulk")
			continue
		}
		fackConn := &connection.Connection{}
		rep := handler.database.Exec(fackConn, r.Args)
		if reply.IsErrorReply(rep) {
			logger.Error(rep)
		}
	}
}
