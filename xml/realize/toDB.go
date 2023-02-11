package realize

import (
	"encoding/xml"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"strings"
)

type Contacts struct {
	C []Contact `xml:"contact"`
}

type Contact struct {
	Name  string `xml:"name"`
	Sex   string `xml:"sex,attr"`
	Phone string `xml:"phone"`
}

// ToDB 将数据从XML文件写到数据库
func ToDB(db *gorm.DB, filePath string) error {
	//读入文件并反序列化
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var c = Contacts{}
	err1 := xml.Unmarshal(file, &c)
	if err1 != nil {
		return err1
	}
	//fmt.Println(c)

	//将数据写到对应的数据库struct中
	var contacts = make([]ContactsModel, 0, len(c.C))
	for _, v := range c.C {
		contacts = append(contacts,
			ContactsModel{
				Name:  v.Name,
				Sex:   v.Sex,
				Phone: v.Phone,
			})
	}
	//fmt.Println(contacts)

	//将数据写到数据库
	db.Model(&ContactsModel{}).Create(&contacts)

	return nil
}

// DoToDB 改进后的ToDB
func DoToDB(db *gorm.DB, filePath string) error {
	//读入文件并反序列化
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	var c = Contacts{}
	err1 := xml.Unmarshal(file, &c)
	if err1 != nil {
		return err1
	}
	//fmt.Println(c)

	//进行数据的校验
	//这里将读入文件中含有相同电话号码视为数据有错误，读入的数据和数据库中数据有相同的电话号码视为修改
	//校验文件
	m := make(map[string]bool)
	for _, v := range c.C {
		phone := v.Phone
		if m[phone] == true {
			return fmt.Errorf("输入的数据有误")
		}
		m[phone] = true
	}
	//校验数据库中是否有相同的电话号码
	fromDB := FromDB(db)
	for _, v := range fromDB {
		//复用m
		if m[v.Phone] == true { //这里将需要改变的标记为false
			m[v.Phone] = false
		}
	}

	//将需要创建的值和更改的值放到不同的slice中
	//这里还需要将从XML文件中读入的数据转存到数据库对应的结构体中
	createSlice := make([]ContactsModel, 0)
	updateSlice := make([]ContactsModel, 0)
	for _, v := range c.C {
		temp := ContactsModel{
			Name:  replaceString(v.Name),
			Sex:   replaceString(v.Sex),
			Phone: replaceString(v.Phone),
		}
		if m[v.Phone] == true {
			createSlice = append(createSlice, temp)
			continue
		}
		updateSlice = append(updateSlice, temp)
	}
	//fmt.Println(createSlice, updateSlice)

	//利用gorm执行创建和修改操作
	//if len(createSlice) != 0 {
	//	db.Model(&ContactsModel{}).Create(&createSlice)
	//}
	//for _, v := range updateSlice {
	//	db.Model(&ContactsModel{}).Where("phone=?", v.Phone).Updates(v)
	//}
	//打开数据库事务
	db.Transaction(func(tx *gorm.DB) error {
		if len(createSlice) != 0 {
			result := tx.Model(&ContactsModel{}).Create(&createSlice)
			if result.Error != nil {
				return result.Error
			}
		}
		for _, v := range updateSlice {
			result := tx.Model(&ContactsModel{}).Where("phone=?", v.Phone).Updates(v)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})

	return nil
}

//去除字符串中的空格和换行符
func replaceString(s string) (newString string) {
	s = strings.Replace(s, "\n", "", -1)
	newString = strings.Replace(s, " ", "", -1)
	return
}
