package realize

import (
	"encoding/xml"
	"gorm.io/gorm"
	"io/ioutil"
)

type ContactsModel struct { //定义与数据库字段相对应的结构体
	gorm.Model
	Name  string
	Sex   string
	Phone string
}

func (*ContactsModel) TableName() string {
	return "go_xml"
}

// FromDB 从数据库读取数据
func FromDB(db *gorm.DB) []ContactsModel {
	contacts := make([]ContactsModel, 0)
	db.Find(&contacts)
	//fmt.Println(contacts)
	return contacts
}

// ToXMLModel 从数据库读到的数据中，将需要的字段拷贝到Contact结构体中
func ToXMLModel(contacts []ContactsModel) Contacts {
	c := make([]Contact, 0, len(contacts)) //我们可以确定返回的[]Contact的长度和传入的slice的长度是一样的，这里我们初始化切片的时候就可以指定长度，而不是每次append的时候再去申请，这样可以提高性能
	for _, v := range contacts {
		temp := Contact{
			Name:  v.Name,
			Sex:   v.Sex,
			Phone: v.Phone,
		}
		c = append(c, temp)
	}
	//fmt.Println(c)
	return Contacts{
		c,
	}
}

// ToXML 将数据序列化后写到文件
func ToXML(c *Contacts) error {
	//data, err := xml.Marshal(c)//使用这个序列化XML文件会全部写在一行
	//fmt.Println(c)
	data, err := xml.MarshalIndent(c, "", "    ") //为XML文件整理格式
	if err != nil {
		return err
	}
	ioutil.WriteFile("./realize/output.xml", data, 0644)
	return nil
}

// DoFromDB 调用前面的函数，完成操作
func DoFromDB(db *gorm.DB) error {
	fromDB := FromDB(db)
	model := ToXMLModel(fromDB)
	err := ToXML(&model) //返回error这里没有做处理
	return err
}
