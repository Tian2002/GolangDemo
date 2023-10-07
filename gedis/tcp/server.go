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
	}
	waitDown.Wait()
}
