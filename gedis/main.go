package main

//gedis/main.go

import (
	"fmt"
	"gedis/config"
	"gedis/lib/logger"
	"gedis/resp/handler"
	"gedis/tcp"
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
		handler.MakeHandler()) //这里是关键，我们需要在这里注册服务，这里暂时是注册的Echo，后面我们只需要把真正的业务注册在这里就可以了
	if err != nil {
		logger.Error(err)
	}
}
