package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"xml/realize"
)

func main() {
	dsn := "root:2515426141@tcp(127.0.0.1:3306)/xml?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败")
	}
	//先将XMl文件的数据写入数据库，再将数据从数据库读出来
	err1 := realize.DoToDB(db, "./realize/input.xml")
	if err != nil {
		panic(err1)
	}
	err2 := realize.DoFromDB(db)
	if err2 != nil {
		panic(err2)
	}
}
