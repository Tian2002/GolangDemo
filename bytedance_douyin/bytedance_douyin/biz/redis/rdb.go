package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

//var Rdb *redis.Client

var Ctx = context.Background()
var RdbUserLike *redis.Client
var RdbVideoLike *redis.Client
var RdbUserDoNotLike *redis.Client

// InitRDB 创建redis client
func InitRDB() {
	rdbIp := "127.0.0.1"
	//rdbIp := config.Config.Redis.RdbIP
	rdbPort := "6379"
	//rdbPort := config.Config.Redis.RdbPort
	RdbUserLike = redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", rdbIp, rdbPort),
			Password: "", //没有设置密码
			DB:       0,
		})
	RdbVideoLike = redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", rdbIp, rdbPort),
			Password: "", //没有设置密码
			DB:       1,
		})
	RdbUserDoNotLike = redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", rdbIp, rdbPort),
			Password: "", //没有设置密码
			DB:       2,
		})
}
