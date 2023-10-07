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
