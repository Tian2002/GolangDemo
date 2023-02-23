package utils

import (
	"time"
)

// Fn 定义函数类型
type Fn func() error

// MyTicker 定义一个结构体
type MyTicker struct {
	MyTick *time.Ticker
	Runner Fn
}

// NewTick  新建一个定时任务的结构体，传入一个x秒，一个要执行的函数
func NewTick(interval int, f Fn) *MyTicker {
	return &MyTicker{
		MyTick: time.NewTicker(time.Duration(interval) * time.Second),
		Runner: f,
	}
}

// Start 启动定时器需要执行的任务
func (t *MyTicker) Start() error {
	for {
		select {
		case <-t.MyTick.C:
			err := t.Runner()
			if err != nil {
				return err
			}
		}
	}
}
