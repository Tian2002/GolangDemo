package config

import (
	"bytedance_douyin/biz/dao"
	"bytedance_douyin/biz/model/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	dsn := "root:2515426141@tcp(127.0.0.1:3306)/bytedance_douyin?charset=utf8mb4&parseTime=True&loc=Local"
	//dsn := Config.Mysql.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Database connection error")
	}
	*dao.DB = *db
	db.AutoMigrate(&entity.User{}, &entity.Follow{}, &entity.Video{}, &entity.UserInfo{}, &entity.Favorite{}, &entity.Comment{})
}
