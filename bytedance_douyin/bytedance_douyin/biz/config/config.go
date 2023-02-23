package config

import (
	"github.com/cloudwego/hertz/pkg/common/json"
	"io/ioutil"
)

type File struct {
	Mysql struct {
		Dsn string `json:"dsn"`
	} `json:"mysql"`
	Redis struct {
		RdbIP   string `json:"rdbIp"`
		RdbPort string `json:"rdbPort"`
	} `json:"redis"`
	JWT struct {
		StSigningKey string `json:"stSigningKey"`
		ExpiresTime  int    `json:"ExpiresTime"`
	} `json:"JWT"`
	MD5 struct {
		Salt string `json:"salt"`
	} `json:"MD5"`
	Ffmpeg struct {
		ExePath string `json:"exePath"`
	} `json:"ffmpeg"`
	Feed struct {
		MaxVideos int `json:"MaxVideos"`
	} `json:"feed"`
	App struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"app"`
}

var Config File = File{}

func SetConfig() {
	readFile, err := ioutil.ReadFile("./biz/config/config.json")
	if err != nil {
		panic("读取配置文件失败" + err.Error())
	}
	err = json.Unmarshal(readFile, &Config)
	if err != nil {
		panic("反序列化配置文件失败")
	}
}
