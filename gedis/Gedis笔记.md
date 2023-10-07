# Gedis笔记



## 前置工具配置



### 配置日志

~~~go
package logger

///gedis/lib/logger/flle.go
import (
	"fmt"
	"os"
)

func checkNotExist(src string) bool {
	_, err := os.Stat(src)    //判断文件属性
	return os.IsNotExist(err) //文件是否不存在
}

func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

func isNotExistMkDir(src string) error {
	if notExist := checkNotExist(src); notExist == true {
		if err := mkDir(src); err != nil {
			return err
		}
	}
	return nil
}

func mkDir(src string) error {
	err := os.MkdirAll(src, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func mustOpen(fileName, dir string) (*os.File, error) {
	perm := checkPermission(dir)
	if perm == true {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}

	err := isNotExistMkDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error during make dir %s, err: %s", dir, err)
	}

	f, err := os.OpenFile(dir+string(os.PathSeparator)+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("fail to open file, err: %s", err)
	}

	return f, nil
}

~~~



~~~go
package logger

///gedis/lib/logger/logger.go
import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Settings stores config for logger
type Settings struct {
	Path       string `yaml:"path"`
	Name       string `yaml:"name"`
	Ext        string `yaml:"ext"`
	TimeFormat string `yaml:"time-format"`
}

var (
	logFile            *os.File
	defaultPrefix      = ""
	defaultCallerDepth = 2
	logger             *log.Logger
	mu                 sync.Mutex
	logPrefix          = ""
	levelFlags         = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

type logLevel int

// log levels
const (
	DEBUG logLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

const flags = log.LstdFlags

func init() {
	logger = log.New(os.Stdout, defaultPrefix, flags)
}

// Setup initializes logger
func Setup(settings *Settings) {
	var err error
	dir := settings.Path
	fileName := fmt.Sprintf("%s-%s.%s",
		settings.Name,
		time.Now().Format(settings.TimeFormat),
		settings.Ext)

	logFile, err := mustOpen(fileName, dir)
	if err != nil {
		log.Fatalf("logging.Setup err: %s", err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(mw, defaultPrefix, flags)
}

func setPrefix(level logLevel) {
	_, file, line, ok := runtime.Caller(defaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d] ", levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s] ", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}

// Debug prints debug log
func Debug(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(DEBUG)
	logger.Println(v...)
}

// Info prints normal log
func Info(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(INFO)
	logger.Println(v...)
}

// Warn prints warning log
func Warn(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(WARNING)
	logger.Println(v...)
}

// Error prints error log
func Error(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(ERROR)
	logger.Println(v...)
}

// Fatal prints error log then stop the program
func Fatal(v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	setPrefix(FATAL)
	logger.Fatalln(v...)
}

~~~



### config文件配置

利用反射读取

~~~go
package config

//gedis/config/config.go
import (
	"bufio"
	"gedis/lib/logger"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ServerProperties defines global config properties
type ServerProperties struct {
	Bind           string `cfg:"bind"`
	Port           int    `cfg:"port"`
	AppendOnly     bool   `cfg:"appendOnly"`
	AppendFilename string `cfg:"appendFilename"`
	MaxClients     int    `cfg:"maxclients"`
	RequirePass    string `cfg:"requirepass"`
	Databases      int    `cfg:"databases"`

	Peers []string `cfg:"peers"`
	Self  string   `cfg:"self"`
}

// Properties holds global config properties
var Properties *ServerProperties

func init() {
	// default config
	Properties = &ServerProperties{
		Bind:       "127.0.0.1",
		Port:       6379,
		AppendOnly: false,
	}
}

func parse(src io.Reader) *ServerProperties {
	config := &ServerProperties{}

	// read config file
	rawMap := make(map[string]string)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}
		pivot := strings.IndexAny(line, " ")
		if pivot > 0 && pivot < len(line)-1 { // separator found
			key := line[0:pivot]
			value := strings.Trim(line[pivot+1:], " ")
			rawMap[strings.ToLower(key)] = value
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Fatal(err)
	}

	// parse format
	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)
	n := t.Elem().NumField()
	for i := 0; i < n; i++ {
		field := t.Elem().Field(i)
		fieldVal := v.Elem().Field(i)
		key, ok := field.Tag.Lookup("cfg")
		if !ok {
			key = field.Name
		}
		value, ok := rawMap[strings.ToLower(key)]
		if ok {
			// fill config
			switch field.Type.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Int:
				intValue, err := strconv.ParseInt(value, 10, 64)
				if err == nil {
					fieldVal.SetInt(intValue)
				}
			case reflect.Bool:
				boolValue := "yes" == value
				fieldVal.SetBool(boolValue)
			case reflect.Slice:
				if field.Type.Elem().Kind() == reflect.String {
					slice := strings.Split(value, ",")
					fieldVal.Set(reflect.ValueOf(slice))
				}
			}
		}
	}
	return config
}

// SetupConfig read config file and store properties into Properties
func SetupConfig(configFilename string) {
	file, err := os.Open(configFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	Properties = parse(file)
}

~~~



### 带超时设置的WaitGroup

~~~go
package wait

///gedis/lib/sync/wait.go
import (
	"sync"
	"time"
)

// Wait 带超时设置的WaitGroup
type Wait struct {
	wg sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

func (w *Wait) Down() {
	w.wg.Done()
}

// WaitWithTimeout 正常Wait结束返回true，超时返回false
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		w.wg.Wait()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

~~~





### 实现atomic的bool

atomic包下是有bool类型的实现的，但是项目源码中是自己实现的，现在我们自己也尝试实现一个bool类型的原子性操作，但是我在项目中使用的是系统包装的，系统中也是包装的uint32实现的bool

~~~go
package atomic

///gedis/lib/sync/atomic.go
import "sync/atomic"

// Boolean 原子性的bool类型
type Boolean uint32

func (b *Boolean) Get() bool {
	return atomic.LoadUint32((*uint32)(b)) != 0
}

func (b *Boolean) Set(val bool) {
	if val {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
~~~



## 实现TCP服务器



### TCP服务器

tcp服务器的逻辑很简单，通过net.Listen监听端口的tcp连接，使用Accept方法接收一个个连接，然后将连接注册到业务中

这里面需要注意的是采用了os.Signal监听系统的信号，来关闭服务器

~~~go
package tcp

//gedis/tcp/server
import (
	"context"
	"gedis/interface/tcp"
	"gedis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct { //启动tcp server的一些配置
	Address string //监听的地址
}

func ListenAndServeWithSignal(config *Config, handler tcp.Handler) error {
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		logger.Info("监听端口", config.Address, "失败")
	}
	logger.Info("开始监听端口", config.Address)

	closeChan := make(chan struct{}) //用来感知系统信号，当系统将程序关闭时，去接收一个信号
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT) //当系统向程序发信号时转发给这个sigChan
	go func() {
		sig := <-sigChan
		logger.Info("系统向程序发送了一个关闭信号：", sig)
		closeChan <- struct{}{}
	}()

	ListenAndServer(listener, handler, closeChan)
	return nil
}

func ListenAndServer(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	go func() {
		<-closeChan
		logger.Info("正在关闭程序")
		listener.Close()
		handler.Close()
	}()

	defer func() { //正常情况下应该在执行完成后关闭连接和业务,系统直接关闭程序时走不到这一步
		listener.Close()
		handler.Close()
	}()

	ctx := context.Background()
	waitDown := sync.WaitGroup{} //用来处理当程序退出时，应该等待已有的连接处理完业务再退出
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Info(err)
			break
		}
		logger.Info("开始监听")
		waitDown.Add(1)
		go func() {
			defer waitDown.Done()
			handler.Handle(ctx, conn)
		}()
		waitDown.Wait()
	}
}

~~~



### 一个简单的业务Echo

实现一个简单的业务Echo，直接将客服端发来的数据原封不动的返回

~~~go
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
		client.Waiting.Down() //业务完成，可以关掉
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

~~~



### TCP服务器测试

~~~go
package main

//gedis/main.go

import (
	"fmt"
	"gedis/config"
	"gedis/lib/logger"
	"gedis/tcp"
	EchoHandler "gedis/tcp"
	"os"
)

// 配置文件的路径
const configFile string = "redis.conf"

// 默认配置
var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

// 查看文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	//日志配置
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "gedis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})
	//配置文件配置
	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}
	//监听端口开启服务
	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d",
				config.Properties.Bind,
				config.Properties.Port),
		},
		EchoHandler.MakeHandler()) //这里是关键，我们需要在这里注册服务，这里暂时是注册的Echo，后面我们只需要把真正的业务注册在这里就可以了
	if err != nil {
		logger.Error(err)
	}
}

~~~

![image-20230924174823986](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230924174823986.png)

注意测试的时候，在后面加上'\n',因为我们默认将换行符作为一条条命令的分割信号

## 实现redis协议解析器RESP



### RESP介绍

+ 正常回复，单字符串

  以“+”开头，“\r\n”结尾

+ 错误恢复

  以“-”开头，“\r\n”结尾

+ 整数

  以“:”开头，“\r\n”结尾

+ 多行字符串

  以“$”开头，中间跟字节数,“\r\n”结尾

  发送“xzh“，即“$3\r\nxzh\r\n”

+ 数组

  以“*”开头，后面跟成员的个数，“\r\n”结尾

  发送“set key value“，即“*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$5\r\nvalue\r\n”

### RESP实现

RESP实现分为两部分，一是将客服端的字节流解析为命令，而是将服务端发送的数据转换为RESP协议对应的字节流

####  Reply 应该回复给客户端的字节流

需要回复给客服端的字节流的创建方式还是很简单的

以常见的"PONG"为例，

![image-20230924180054598](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230924180054598.png)

我们只需要创建一个结构体，写一个ToBytes方法，按照RESP规则将需要返回的数据封装进去就行了。

我们还常常创建一个对应的Make方法，作为一个该结构体的构造器，这里第9行的处理可以节约内存，但是这样的处理仅对这样的常量返回有用，对于一些需要携带一些具体数据的reply不能这样处理，如

![image-20230924180609829](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230924180609829.png)

在这里，我们具体创建了常量回复、正常回复、标准错误回复这三类回复的数个结构体与方法

~~~go
package reply

//gedis/resp/reply/consts.go
//固定的回复

type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")
var thePongReply = new(PongReply) //在本地持有一个该类型的指针，避免每次都创建一个该类型的指针，节约内存
func (r PongReply) ToBytes() []byte {
	return pongBytes
}
func MakePongReply() *PongReply {
	return thePongReply
}

type OKReply struct{}

var OKBytes = []byte("+OK\r\n")
var theOkReply = new(OKReply)

func (r OKReply) ToBytes() []byte {
	return OKBytes
}
func MakeOKReply() *OKReply {
	return theOkReply
}

// NullBulkReply 空的字符串回复,是nil，而不是""
type NullBulkReply struct {
}

var nullBulkBytes = []byte("$-1\r\n")
var theNullBulkReply = new(NullBulkReply)

func (r NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}
func MakeNullBulkReply() *NullBulkReply {
	return theNullBulkReply
}

// EmptyMultiBulkReply 这里是返回空数组
type EmptyMultiBulkReply struct {
}

var emptyMultiBulkBytes = []byte("*0\r\n")
var theEmptyMultiBulkReply = new(EmptyMultiBulkReply)

func (r EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}
func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return theEmptyMultiBulkReply
}

// NoReply 回复一个真的空
type NoReply struct{}

var noReplyBytes = []byte("")
var theNoReply = new(NoReply)

func (r NoReply) ToBytes() []byte {
	return noReplyBytes
}
func MakeNoReply() *NoReply {
	return theNoReply
}

~~~

~~~go
package reply

//gedis/resp/reply/reply.go

import (
	"bytes"
	"gedis/interface/resp"
	"strconv"
)

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

var (
	nullBulkReplyBytes = []byte("$-1\r\n")
	CRLF               = "\r\n"
)

// BulkReply 返回多行字符串，以“$”开头，中间跟字节数,“\r\n”结尾
// 发送“xzh“，即“$3\r\nxzh\r\n”
type BulkReply struct {
	Arg []byte
}

func (r *BulkReply) ToBytes() []byte {
	l := len(r.Arg)
	if l == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(l) + CRLF + string(r.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// MultiBulkReply 返回数组，以“*”开头，后面跟成员的个数，“\r\n”结尾
// 发送“set key value“，即“*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
type MultiBulkReply struct {
	Args [][]byte
}

func (r *MultiBulkReply) ToBytes() []byte {
	l := len(r.Args)
	var buf bytes.Buffer //这样写效率更高
	buf.WriteString("*" + strconv.Itoa(l) + CRLF)
	for i := 0; i < l; i++ {
		buf.WriteString(string((&BulkReply{Arg: r.Args[i]}).ToBytes()))
	}
	return buf.Bytes()
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

// StatusReply 返回一些简单的状态
type StatusReply struct {
	Status string
}

func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

// IntReply 以“:”开头，“\r\n”结尾
type IntReply struct {
	Code int64
}

func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

// StandardErrReply 返回一些标准的错误
type StandardErrReply struct {
	Status string
}

func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// IsErrorReply 判断这个reply是不是错误
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}

~~~

~~~go
package reply

//gedis/resp/reply/error.go

// UnKnowErrReply 未知错误
type UnKnowErrReply struct{}

var unKnowErrBytes = []byte("-Err unKnow\r\n")
var theUnKnowErrReply = new(UnKnowErrReply)

func (r UnKnowErrReply) Error() string {
	return string(unKnowErrBytes)
}

func (r UnKnowErrReply) ToBytes() []byte {
	return unKnowErrBytes
}

func MakeUnKnowErrReply() *UnKnowErrReply {
	return theUnKnowErrReply
}

// ArgNumErrReply 客户端发送到置零参数个数错误
type ArgNumErrReply struct {
	Cmd string //记录一下指令
}

func (r *ArgNumErrReply) Error() string {
	return string(r.ToBytes())
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-Err wrong number of arguments for'" + r.Cmd + "' command\r\n")
}
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{Cmd: cmd}
}

// SyntaxErrReply 语法错误
type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-Err syntax error\r\n")
var theSyntaxErrReply = new(SyntaxErrReply)

func (r SyntaxErrReply) Error() string {
	return string(syntaxErrBytes)
}

func (r SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

// WrongTypeErrReply 数据类型错误
type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-Err WrongType Operation against a key holding the wrong kind of value\r\n")
var theWrongTypeErrReply = new(WrongTypeErrReply)

func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return string(r.ToBytes())
}
func MakeWrongTypeErrReply() *WrongTypeErrReply {
	return theWrongTypeErrReply
}

// ProtocolErrReply 协议错误
type ProtocolErrReply struct {
	Msg string
}

func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n")
}

func (r *ProtocolErrReply) Error() string {
	return string(r.ToBytes())
}
func MakeProtocolErrReply(msg string) *ProtocolErrReply {
	return &ProtocolErrReply{Msg: msg}
}

~~~



####  解析客服端发来的字节流

解析时主要注意以下几点

+ 转移字符只占一个长度  "\r\n"的长度为2	
+ 当读取"$7\r\n12\r\n567\r\n"这样的数据时，我们不能直接ReadLine，因为中间可能会有"\r\n",这里我们使用io.ReadAll
+ 解析协议时我们使用异步解析，和数据库处理业务分开

~~~go
package parser

//gedis/resp/parser/parser.go
import (
	"bufio"
	"errors"
	"gedis/interface/resp"
	"gedis/lib/logger"
	"gedis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

// Payload 客服端发送来的数据
type Payload struct {
	Data resp.Reply //这是使用服务端定义的结构是因为互相发送的消息的结构是一样的
	Err  error
}

// 解析器的状态
type readState struct {
	readingMultiLine  bool     //在解析单行还是多行数据，如果是多行则还没读取完成，如遇到"*"或"$"
	expectedArgsCount int      //正在读取指令应该有几个参数，字符串数组用的
	msgType           byte     //指令的类型
	args              [][]byte //传过来的解析后的数据
	bulkLen           int64    //解析的字符串的长度
}

// ParseStream 异步的去解析协议，把解析的结构放进一个channel中
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch) //将协议解析去异步执行
	return ch
}

// 计算解析有没有完成
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for { //循环解析，解析每一行
		// read line
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr { // 遇到io错误，停止读取，把已经读到的数据写入channel
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			//协议错误，重置读取的状态
			ch <- &Payload{
				Err: err,
			}
			state = readState{}
			continue
		}

		// 解析这一条的命令
		if !state.readingMultiLine { // 刚开始时，接收新响应或者state.readingMultiLine已经被标记为true
			if msg[0] == '*' { //依次判断，看看是RESP协议中五种数据的哪一种，这里是判断是不是数组
				// 解析multiBulk
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{} // reset state
					continue
				}
			} else if msg[0] == '$' { // 解析多行字符串
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.bulkLen == -1 { // null bulk reply
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{} // reset state
					continue
				}
			} else {
				// 解析正常回复、错误回复、整数回复
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{} // reset state
				continue
			}
		} else {
			// 收到以下批量回复
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{} // reset state
				continue
			}
			// if sending finished
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}
}

// 这个方法是读取一条的命令，第二个参数是标识当error不为空时，错误是否为io的错误
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 { // 只有解析到"$"符号后，state.bulkLen才会变，当正在解析"$"时也还是0
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { // read bulk line (binary safe)
		msg = make([]byte, state.bulkLen+2)  //state.bulkLen只标识了字符串和的长度，所有还要加上"\r\n"的长度
		_, err = io.ReadFull(bufReader, msg) //这里使用ReadFull, 而不继续使用 ReadBytes('\n') 读取下一行, 这是因为在字符串内部可能包含"\r\n",这里应该特殊处理
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0 //解析完成后重置state.bulkLen，因为下一行又是$开头或者直接结束
	}
	return msg, false, nil
}

func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// first line of multi bulk reply
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+': // status reply
		result = reply.MakeStatusReply(str[1:])
	case '-': // err reply
		result = reply.MakeErrReply(str[1:])
	case ':': // int reply
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// 阅读多批量回复或批量回复的非第一行，即1、读到了"*"开头的后面 或 2、先前读到了"$"开头的后面
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' { //对应1
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else { //对应2
		state.args = append(state.args, line)
	}
	return nil
}

~~~

### 测试RESP

和之前的Echo测试类似，这里我们也测试一下

写一个有关客服端连接的文件

```go
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
```

一个具体业务处理的文件，这里就是把发送进入channel的数据写回去，写回去时是调用的reply.MakeMultiBulkReply(args)，这是因为客户端发来的数据一般是数组

```go
package handler

/*
 * A tcp.RespHandler implements redis protocol
 */
//gedis/resp/handler/handler.go
import (
	"context"
	"gedis/database"
	databaseface "gedis/interface/database"
	"gedis/lib/logger"
	"gedis/resp/connection"
	"gedis/resp/parser"
	"gedis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

// RespHandler 实现接口tcp.Handler作为redis的handler
type RespHandler struct {
	activeConn sync.Map // *client -> placeholder
	db         databaseface.Database
	closing    atomic.Bool // refusing new client and new request
}

func MakeHandler() *RespHandler {
	var db databaseface.Database
	db = database.NewEchoDatabase()
	return &RespHandler{
		db: db,
	}
}

func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Handle 接收并执行redis命令，这里这个业务实现的很简单，就是打印发来的命令
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Load() {
		// closing handler refuse new connection
		_ = conn.Close()
	}

	client := connection.NewConn(conn)
	h.activeConn.Store(client, 1)

	ch := parser.ParseStream(conn)
	for payload := range ch { //把解析的指令输出
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok { //一般客服端的数据都是通过数组发来的，但是其他的我们也能解析，但是这里打印一个错误日志
			logger.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args) //这里模拟的是执行指令，操作数据库
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close stops handler
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Store(true)
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}

```

上面提到了模拟数据库处理，这里写了一个假的数据库实现

```go
package database

//gedis/database/echo_database.go
import (
   "gedis/interface/resp"
   "gedis/lib/logger"
   "gedis/resp/reply"
)

type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
   return &EchoDatabase{}
}

func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
   return reply.MakeMultiBulkReply(args)

}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
   logger.Info("EchoDatabase AfterClientClose")
}

func (e EchoDatabase) Close() {
   logger.Info("EchoDatabase Close")

}
```

![image-20230924215407547](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230924215407547.png)



## 实现内存数据库

实现redis内核database



### Dict

实现一个dict接口，当我们换底层实现时，不必修改上层的命令，比如从map+读写锁到sync.Map

```go
package dict

//gedis/datastruct/dict/dict.go

// Consumer 遍历的时候传入的方法，和sync,Map的range方法的参数类似
type Consumer func(key string, val interface{}) bool

// Dict 实现一个dict接口，当我们换底层实现时，不必修改上层的命令，比如从map+读写锁到sync.Map
type Dict interface {
   Get(key string) (val interface{}, exists bool)
   // Len 返回字典里面有多少数据
   Len() int
   Put(key string, val interface{}) (result int)
   // PutIfAbsent 如果没有才进行操作
   PutIfAbsent(key string, val interface{}) (result int)
   // PutIfExists 如果有才进行操作
   PutIfExists(key string, val interface{}) (result int)
   Remove(key string) (result int)
   ForEach(consumer Consumer)
   // Keys 返回所有键
   Keys() []string
   // RandomKeys 随机返回键
   RandomKeys(limit int) []string
   // RandomDistinctKeys 随机返回不重复的键
   RandomDistinctKeys(limit int) []string
   clear()
}
```

实现syncDict，即该接口的具体实现

```go
package dict

//gedis/datastruct/dict/sync_dict.go
import "sync"

type SyncDict struct {
   m sync.Map
}

func MakeSyncDict() *SyncDict {
   return &SyncDict{}
}

func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
   return dict.m.Load(key)
}

func (dict *SyncDict) Len() int {
   l := 0
   dict.m.Range(func(key, value any) bool {
      l++
      return true
   })
   return l
}

func (dict *SyncDict) Put(key string, val interface{}) (result int) {
   _, exists := dict.m.Load(key)
   dict.m.Store(key, val)
   if exists { //原本存在key返回0，而不是失败了
      return 0
   }
   return 1
}

// PutIfAbsent 没有这个key才操作
func (dict *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
   _, exists := dict.m.Load(key)
   if !exists {
      dict.m.Store(key, val)
      return 1
   }
   return 0
}

// PutIfExists 存在这个key才操作
func (dict *SyncDict) PutIfExists(key string, val interface{}) (result int) {
   _, exists := dict.m.Load(key)
   if exists {
      dict.m.Store(key, val)
      return 1
   }
   return 0
}

func (dict *SyncDict) Remove(key string) (result int) {
   _, exists := dict.m.Load(key)
   if exists {
      dict.m.Delete(key)
      return 1
   }
   return 0
}

func (dict *SyncDict) ForEach(consumer Consumer) {
   dict.m.Range(func(key, value any) bool {
      consumer(key.(string), value)
      return true //这里一直返回true，让他施加到所有的k，v上
   })
}

func (dict *SyncDict) Keys() []string {
   res := make([]string, dict.Len())
   i := 0
   dict.m.Range(func(key, value any) bool {
      res[i] = key.(string)
      i++
      return true
   })
   return res
}

func (dict *SyncDict) RandomKeys(limit int) []string {
   res := make([]string, limit)
   for i := 0; i < limit; i++ {
      dict.m.Range(func(key, value any) bool {
         res[i] = key.(string)
         return false //每次去随机取一个
      })
   }
   return res
}

func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
   res := make([]string, limit)
   i := 0
   //一次随机取limit个
   dict.m.Range(func(key, value any) bool {
      res[i] = key.(string)
      i++
      return i < limit
   })
   return res
}

func (dict *SyncDict) clear() {
   *dict = *MakeSyncDict()
}
```

### 数据库的实现



#### 上层命令调用逻辑

将全部指令包装为一个个结构体，放到map中，通过指令名称查找

```go
package database

//gedis/database/command.go
import "strings"

var cmdTable = make(map[string]*command) //记录系统所有的指令名称和对应的command

type command struct {
   exector ExecFunc //具体的执行函数
   arity   int      //参数的数量
}

// RegisterCommand 注册command
func RegisterCommand(name string, exector ExecFunc, arity int) {
   name = strings.ToLower(name) //命令全部转换为小写，避免大小写的影响
   cmdTable[name] = &command{
      exector: exector,
      arity:   arity,
   }
}
```

实现数据库对每一个指令的调用逻辑

```go
package database

//gedis/database/db.go
import (
   "gedis/datastruct/dict"
   "gedis/interface/resp"
   "gedis/resp/reply"
   "strings"
)

type DB struct {
   index int
   data  dict.Dict
}

// ExecFunc 方法执行函数
type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine = [][]byte

func MakeDB() *DB {
   return &DB{data: dict.MakeSyncDict()}
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
```

上面每个ExecFunc具体的逻辑待实现



#### 每个指令内部具体逻辑的实现

```go
package database

//gedis/database/ping.go
import (
   "gedis/interface/resp"
   "gedis/resp/reply"
)

// Ping 实现数据库内的操作以及相对应的返回
func Ping(db *DB, args [][]byte) resp.Reply {
   return reply.MakePongReply()
}

// 还需要在程序启动时注册Ping指令
func init() {
   RegisterCommand("ping", Ping, 1)
}
```

```go
package database

//gedis/database/keys.go
import (
   "gedis/interface/resp"
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
   return reply.MakeIntReply(1)
}

// FLUSH
func execFlushDB(db *DB, args [][]byte) resp.Reply {
   db.Flush()
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
```

```string
package database

//gedis/database/string.go
import (
   "gedis/interface/database"
   "gedis/interface/resp"
   "gedis/resp/reply"
)

func (db *DB) getAsString(key string) ([]byte, reply.ErrorReply) {
   entity, ok := db.GetEntity(key)
   if !ok {
      return nil, nil
   }
   bytes, ok := entity.Data.([]byte)
   if !ok {
      return nil, &reply.WrongTypeErrReply{}
   }
   return bytes, nil
}

// execGet GET KEY
func execGet(db *DB, args [][]byte) resp.Reply {
   key := string(args[0])
   bytes, err := db.getAsString(key)
   if err != nil {
      return err
   }
   if bytes == nil {
      return &reply.NullBulkReply{}
   }
   return reply.MakeBulkReply(bytes)
}

// execSet SET KEY
func execSet(db *DB, args [][]byte) resp.Reply {
   key := string(args[0])
   value := args[1]
   entity := &database.DataEntity{
      Data: value,
   }
   db.PutEntity(key, entity)
   return reply.MakeOKReply()
}

// execSetNX
func execSetNX(db *DB, args [][]byte) resp.Reply {
   key := string(args[0])
   value := args[1]
   entity := &database.DataEntity{
      Data: value,
   }
   result := db.PutIfAbsent(key, entity)
   return reply.MakeIntReply(int64(result))
}

// execGetSet
func execGetSet(db *DB, args [][]byte) resp.Reply {
   key := string(args[0])
   value := args[1]

   entity, exists := db.GetEntity(key)
   db.PutEntity(key, &database.DataEntity{Data: value})
   if !exists {
      return reply.MakeNullBulkReply()
   }
   old := entity.Data.([]byte)
   return reply.MakeBulkReply(old)
}

// execStrLen 返回数据的长度
func execStrLen(db *DB, args [][]byte) resp.Reply {
   key := string(args[0])
   entity, exists := db.GetEntity(key)
   if !exists {
      return reply.MakeNullBulkReply()
   }
   old := entity.Data.([]byte)
   return reply.MakeIntReply(int64(len(old)))
}

func init() {
   RegisterCommand("Get", execGet, 2)
   RegisterCommand("Set", execSet, -3)
   RegisterCommand("SetNx", execSetNX, 3)
   RegisterCommand("GetSet", execGetSet, 3)
   RegisterCommand("StrLen", execStrLen, 2)
}
```

#### 数据库的创建与指令传递调用

之前我们写了一个echodatabase，现在我们要实现一个真正的database

```go
package database

//gedis/database/database.go
import (
   "fmt"
   "gedis/config"
   "gedis/interface/resp"
   "gedis/lib/logger"
   "gedis/resp/reply"
   "runtime/debug"
   "strconv"
   "strings"
)

// Database 在Redis中默认是有16个db的，这里就是将多个数据库记录下来
type Database struct {
   dbSet []*DB
}

// NewDatabase 创建每个数据库
func NewDatabase() *Database {
   mdb := &Database{}
   if config.Properties.Databases == 0 {
      config.Properties.Databases = 16
   }
   mdb.dbSet = make([]*DB, config.Properties.Databases)
   for i := range mdb.dbSet {
      singleDB := makeDB()
      singleDB.index = i
      mdb.dbSet[i] = singleDB
   }
   return mdb
}

// Exec executes command
// 参数“cmdLine”包含命令及其参数，例如：“set key value”
func (mdb *Database) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
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
func (mdb *Database) Close() {
}

func (mdb *Database) AfterClientClose(c resp.Connection) {
}

// 执行选择数据库的逻辑
func execSelect(c resp.Connection, mdb *Database, args [][]byte) resp.Reply {
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
```

然后我们会到mian.go中，这里注册的服务已经修改了，

![image-20230926142403175](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230926142403175.png)

点进去，将我们原来注册的echodb换成我们新写的db

![image-20230926142452109](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230926142452109.png)

一个简单的测试

![image-20230926142606101](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230926142606101.png)

## 小结

到这里我们已经实现了一个简单的单机redis，当然了，我们支持的命令现在还只有string类型相关的简单命令已经select、flush等



## 实现数据库的持久化（AOF）

思路

+ 将和数据库相关的写命令追加近aof文件，按照resp协议写入
+ 将aof文件中的数据读出来写入内存，我们是按照resp协议写入的，我们模拟这些数据是从TCP连接拿到的就行了

```
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
```

## 实现简单的redis集群



### 为什么使用一致性哈希

当我们使用简单的哈希函数时

```
m = hash(o) mod n
```

- 其中，o为对象名称，n为机器的数量，m为机器编号。

考虑以下例子：

3个机器节点，10个数据 的哈希值分别为1,2,3,4,…,10。使用的哈希函数为：(`m=hash(o) mod 3`)
机器0 上保存的数据有：3，6，9
机器1 上保存的数据有：1，4，7，10
机器2 上保存的数据有：2，5，8

当增加一台机器后，此时n = 4，各个机器上存储的数据分别为：

```text
机器0 上保存的数据有：4，8
机器1 上保存的数据有：1，5，9
机器2 上保存的数据有：2，6，10
机器3 上保存的数据有：3，7
```

只有数据1和数据2没有移动，所以当集群中数据量很大时，采用一般的哈希函数，在节点数量动态变化的情况下会造成大量的数据迁移，导致网络通信压力的剧增，严重情况，还可能导致数据库宕机。

这时我们就需要使用一致性哈希

~~~go
package consistenthash

//gedis/lib/consistenthash/consistenthash.go

import (
	"hash/crc32"
	"sort"
)

type HashFunc func(data []byte) uint32

type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int          //这里不使用uint32是因为后面会进行排序，我们想要使用go里面自带的排序函数
	nodeHashMap map[int]string //nodeHash->节点
}

func NewNodeMap(fn HashFunc) *NodeMap {
	if fn == nil {
		fn = crc32.ChecksumIEEE //默认的方法
	}
	return &NodeMap{
		hashFunc:    fn,
		nodeHashMap: make(map[int]string),
	}
}

// IsEmpty 判断是否初始化
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

// AddNode 传入某个节点的标识，将他放入节点中。
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodeHashMap[hash] = key
	}
	sort.Ints(m.nodeHashs)
}

// PickNode 判断每个key取那个节点
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	index := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	if index == len(m.nodeHashs) {
		index = 0
	}
	return m.nodeHashMap[m.nodeHashs[index]]
}

~~~



### 简单集群架构

![](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230930162940637.png)

我们在原本的database(在项目中改为standalong_database)上封装一个cluster_database,cluster_database不做业务，只用来做消息的转发

那我们怎么实现节点A到节点B间的通讯呢，比如从A节点到节点B,这是节点A就要作为客户端，我们还可以实现一个连接池去应对高并发的转发



### 简单客户端实现

```go
package client

//gedis/resp/client/client.go
import (
   "gedis/interface/resp"
   "gedis/lib/logger"
   "gedis/lib/sync/wait"
   "gedis/resp/parser"
   "gedis/resp/reply"
   "net"
   "runtime/debug"
   "sync"
   "time"
)

// Client 是管道模式的redis客户端
type Client struct {
   conn        net.Conn
   pendingReqs chan *request // wait to send
   waitingReqs chan *request // waiting response
   ticker      *time.Ticker
   addr        string

   working *sync.WaitGroup // 其计数器显示未完成的请求（挂起和等待）
}

// 请求是发送到redis服务器的消息
type request struct {
   id        uint64
   args      [][]byte
   reply     resp.Reply
   heartbeat bool
   waiting   *wait.Wait
   err       error
}

const (
   chanSize = 256
   maxWait  = 3 * time.Second
)

// MakeClient creates a new client
func MakeClient(addr string) (*Client, error) {
   conn, err := net.Dial("tcp", addr)
   if err != nil {
      return nil, err
   }
   return &Client{
      addr:        addr,
      conn:        conn,
      pendingReqs: make(chan *request, chanSize),
      waitingReqs: make(chan *request, chanSize),
      working:     &sync.WaitGroup{},
   }, nil
}

// Start 启动异步goroutines
func (client *Client) Start() {
   client.ticker = time.NewTicker(10 * time.Second)
   go client.handleWrite()
   go func() {
      err := client.handleRead()
      if err != nil {
         logger.Error(err)
      }
   }()
   go client.heartbeat()
}

// Close 关闭客服端连接的相关资源
func (client *Client) Close() {
   client.ticker.Stop()
   // stop new request
   close(client.pendingReqs)

   // wait stop process
   client.working.Wait()

   // clean
   _ = client.conn.Close()
   close(client.waitingReqs)
}

func (client *Client) handleConnectionError(err error) error {
   err1 := client.conn.Close()
   if err1 != nil {
      if opErr, ok := err1.(*net.OpError); ok {
         if opErr.Err.Error() != "use of closed network connection" {
            return err1
         }
      } else {
         return err1
      }
   }
   conn, err1 := net.Dial("tcp", client.addr)
   if err1 != nil {
      logger.Error(err1)
      return err1
   }
   client.conn = conn
   go func() {
      _ = client.handleRead()
   }()
   return nil
}

func (client *Client) heartbeat() {
   for range client.ticker.C {
      client.doHeartbeat()
   }
}

func (client *Client) handleWrite() {
   for req := range client.pendingReqs {
      client.doRequest(req)
   }
}

// Send 向服务端发送消息
func (client *Client) Send(args [][]byte) resp.Reply {
   request := &request{
      args:      args,
      heartbeat: false,
      waiting:   &wait.Wait{},
   }
   request.waiting.Add(1)
   client.working.Add(1)
   defer client.working.Done()
   client.pendingReqs <- request
   timeout := request.waiting.WaitWithTimeout(maxWait)
   if timeout {
      return reply.MakeErrReply("server time out")
   }
   if request.err != nil {
      return reply.MakeErrReply("request failed")
   }
   return request.reply
}

func (client *Client) doHeartbeat() {
   request := &request{
      args:      [][]byte{[]byte("PING")},
      heartbeat: true,
      waiting:   &wait.Wait{},
   }
   request.waiting.Add(1)
   client.working.Add(1)
   defer client.working.Done()
   client.pendingReqs <- request
   request.waiting.WaitWithTimeout(maxWait)
}

func (client *Client) doRequest(req *request) {
   if req == nil || len(req.args) == 0 {
      return
   }
   re := reply.MakeMultiBulkReply(req.args)
   bytes := re.ToBytes()
   _, err := client.conn.Write(bytes)
   i := 0
   for err != nil && i < 3 {
      err = client.handleConnectionError(err)
      if err == nil {
         _, err = client.conn.Write(bytes)
      }
      i++
   }
   if err == nil {
      client.waitingReqs <- req
   } else {
      req.err = err
      req.waiting.Done()
   }
}

func (client *Client) finishRequest(reply resp.Reply) {
   defer func() {
      if err := recover(); err != nil {
         debug.PrintStack()
         logger.Error(err)
      }
   }()
   request := <-client.waitingReqs
   if request == nil {
      return
   }
   request.reply = reply
   if request.waiting != nil {
      request.waiting.Done()
   }
}

func (client *Client) handleRead() error {
   ch := parser.ParseStream(client.conn)
   for payload := range ch {
      if payload.Err != nil {
         client.finishRequest(reply.MakeErrReply(payload.Err.Error()))
         continue
      }
      client.finishRequest(payload.Data)
   }
   return nil
}
```



### 连接池

使用开源库来建立连接池

~~~shell
 github.com/jolestar/go-commons-pool/v2 v2.1.2 
~~~

这个库我们需要实现工厂模式的接口

```go
package cluster

//gedis/cluster/cluster_pool.go
import (
   "context"
   "errors"
   "gedis/resp/client"
   pool "github.com/jolestar/go-commons-pool/v2"
)

// 实现PooledObjectFactory接口
type connectionFactory struct {
   Peer string //表示连接池连接的那个节点
}

func (f *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
   cli, err := client.MakeClient(f.Peer)
   if err != nil {
      return nil, err
   }
   cli.Start() //登录客服端

   return pool.NewPooledObject(cli), nil
}

func (f *connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
   c, ok := object.Object.(*client.Client)
   if !ok {
      return errors.New("类型错误")
   }
   c.Close()
   return nil
}

//下面几个方法不做具体实现

func (f *connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
   return true
}

func (f *connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
   return nil
}

func (f *connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
   return nil
}
```



### 具体业务

三种执行模式

+ 需要转发到某个集群节点的命令，如set、get
+ 不需要转发的命令，如ping，可以归为1的特殊情况，转发给自己
+ 广播的命令，所有节点都要执行，如flush

```go
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
```



每种不同的命令执行的方法可能不一样，这里我们用一个map[string]CmdFunc来保存

```go
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
```

```go
package cluster

//gedis/cluster/ping.go
import "gedis/interface/resp"

func ping(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply {
   return cluster.db.Exec(c, cmdAndArgs)
}
```

```go
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
```

```go
package cluster

//gedis/cluster/keys.go
import (
   "gedis/interface/resp"
   "gedis/resp/reply"
)

// FlushDB removes all data in current database
func FlushDB(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
   replies := cluster.broadcast(c, args)
   var errReply reply.ErrorReply
   for _, v := range replies {
      if reply.IsErrorReply(v) {
         errReply = v.(reply.ErrorReply)
         break
      }
   }
   if errReply == nil {
      return reply.MakeOKReply()
   }
   return reply.MakeErrReply("error occurs: " + errReply.Error())
}
```

```go
package cluster

//gedis/cluster/del.go
import (
   "gedis/interface/resp"
   "gedis/resp/reply"
)

// Del atomically removes given writeKeys from cluster, writeKeys can be distributed on any node
// if the given writeKeys are distributed on different node, Del will use try-commit-catch to remove them
func Del(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
   replies := cluster.broadcast(c, args)
   var errReply reply.ErrorReply
   var deleted int64 = 0
   for _, v := range replies {
      if reply.IsErrorReply(v) {
         errReply = v.(reply.ErrorReply)
         break
      }
      intReply, ok := v.(*reply.IntReply)
      if !ok {
         errReply = reply.MakeErrReply("error")
      }
      deleted += intReply.Code
   }

   if errReply == nil {
      return reply.MakeIntReply(deleted)
   }
   return reply.MakeErrReply("error occurs: " + errReply.Error())
}
```

```go
package cluster

//gedis/cluster/select.go
import "gedis/interface/resp"

func execSelect(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply {
   return cluster.db.Exec(c, cmdAndArgs)
}
```



集群数据库的业务函数

```go
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
```



修改handler的注册

![image-20230930193942181](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230930193942181.png)



### 小结 

现在我们已经利用go完成了简单的redis

这么我们简单测试一下，我们往6380端口里面写入一些数据，发现 部分数据写入了6379端口对应的aof文件中

![image-20230930195239981](C:\Users\tianpengfei\AppData\Roaming\Typora\typora-user-images\image-20230930195239981.png)
