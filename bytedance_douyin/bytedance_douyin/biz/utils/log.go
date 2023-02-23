package utils

import (
	"log"
	"os"
)

func InitLog() {
	file, err := os.OpenFile("./log/douyin.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal("初始化log失败")
	}
	log.SetOutput(file)
}
