### 使用Golang实现XML文件与数据库间的转换



#### 实现功能

+ 通过Golang编程实现对“通讯录”的XML 文档的解析，并把解析结果存到数据库的表中。
+ 进一步实现相反的过程，即将数据库表的内容读出来，并将其转化为XML 文件存储起来

#### 目的

+ 通过实践去熟悉Gorm的使用
+ 这次使用的XML文件格式很简单，主要是为了熟悉Gorm的增删改查
+ 数据库的导出文件后面也会放在项目文件下

#### 具体实现过程

##### 准备工作

+ 数据库建表

+ 安装Gorm并连接数据库

  在项目打开终端安装Gorm

  ~~~shell
  go get -u gorm.io/gorm
  go get -u gorm.io/driver/sqlite
  ~~~

  安装成功后的go.mod文件

  ![image-20230209125521784](C:\Users\安若浅溪\AppData\Roaming\Typora\typora-user-images\image-20230209125521784.png)

+ 连接数据库

  ~~~go
  dsn := "root:2515426141@tcp(127.0.0.1:3306)/xml?charset=utf8mb4&parseTime=True&loc=Local"
  	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
  	if err != nil {
  		panic("数据库连接失败")
  	}
  ~~~

  

+ 建立与xml文件及数据库表相对应的结构体

  xml对应的结构体

  ~~~go
  type Contacts struct {
  	C []Contact `xml:"contact"`
  }
  
  type Contact struct {
  	Name  string `xml:"name"`
  	Sex   string `xml:"sex,attr"`
  	Phone string `xml:"phone"`
  }
  ~~~

  数据库表对应的结构体 

  ~~~go
  type ContactsModel struct { //定义与数据库字段相对应的结构体
  	gorm.Model
  	Name  string
  	Sex   string
  	Phone string
  }
  
  func (*ContactsModel) TableName() string {
  	return "go_xml"
  }
  ~~~

  这里使用的gorm.Model是gorm库提供的，放在这里其实没有必要，可以直接改为主键ID

  这里与数据库对应的结构体实现了TableName接口，如果不实现该接口，gorm会默认使用该结构体的蛇形复数（contact_models）做为表名。

  在这两个结构体中，字段名的首字母一定要大写，否则可能会导致写入数据失败（因为在Golang中首字母小写表示这个属性只能在同一个包中能被看见）

##### 将数据从数据库到XML文件

+ 将所有数据利用gorm读到一个slice中

  ~~~go
  func FromDB(db *gorm.DB) []ContactsModel {
  	contacts := make([]ContactsModel, 0)
  	db.Find(&contacts)
  	fmt.Println(contacts)
  	return contacts
  }
  ~~~

  

+ 使用range迭代数据，将每个数据中需要的字段存储到Contact的slice中最后全部放到一个Contacts中

  ~~~go
  func ToXMLModel(contacts []ContactsModel) Contacts {
  	c := make([]Contact, 0, len(contacts))
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
  ~~~

  

+ 序列化，存储到文件中

  ~~~go
  func ToXML(c *Contacts) error {
  	//data, err := xml.Marshal(c)
  	//fmt.Println(c)
  	data, err := xml.MarshalIndent(c, "", "    ")
  	if err != nil {
  		return err
  	}
  	ioutil.WriteFile("./realize/output.xml", data, 0644)
  	return nil
  }
  ~~~

+ 相对于原版的使用Java的dom4j转换，不需要自己一个个去建立节点，只需要定义好对应的结构体就可以直接序列化存储到XML文件中

##### 将数据从XML文件到数据库

+ 将数据从XML文件写到一个内存中

  使用ioutil库和XML库的Unmarshal方法

  ~~~go
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
  ~~~

  

+ 将数据写到对应的数据库struct中

  ~~~go
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
  ~~~

  结构体的其余字段gorm会自动填充（deleted_at的值默认为null）

+ 将数据写到数据库

  ~~~go
  db.Model(&ContactsModel{}).Create(&contacts)
  ~~~

  

##### 小结

+ 到这里，我们的实现已经初步完成了，这是我们在main方法中调用函数，将input.xml文件先写到数据库，在将文件读出来。

  ~~~go
  func main() {
  	dsn := "root:2515426141@tcp(127.0.0.1:3306)/xml?charset=utf8mb4&parseTime=True&loc=Local"
  	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
  	if err != nil {
  		panic("数据库连接失败")
  	}
  	realize.ToDB(db, "./realize/input.xml")
  	realize.DoFromDB(db)
  }
  ~~~

  去查询数据库表已经查看新生成的realize/output.xml文件功能已经初步实现成功。

  

+ 但是还有一些问题，例如有两个相同的电话号码存入，这样显然是要经过处理的，这个就需要用到gorm的改操作。

##### 改进将数据从XML文件到数据库

+ 第一步仍然是读入文件并反序列化

  ~~~go
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
  ~~~

  

+ 进行数据的检验

  这里将读入文件中含有相同电话号码视为数据有错误，读入的数据和数据库中数据有相同的电话号码视为修改

  - 检验读入文件中的数据

    将一个个数据放在map中，如果该电话号码没有被标记为true则标记为true，如果被标记过则直接返回error

    ~~~go
    //校验文件
    	m := make(map[string]bool)
    	for _, v := range c.C {
    		phone := v.Phone
    		if m[phone] == true {
    			return fmt.Errorf("输入的数据有误")
    		}
    		m[phone] = true
    	}
    ~~~

    

  - 对比读入文件中的数据与数据库中的数据

    执行到了这一步，在上面的检验中，所有的phone都被记录到map中并标记为true，现在我们只需要拿到数据库中的数据，如果数据中的phone字段在map中被标记为true，则将该电话标记改为false用来和需要重新创建的数据作为区别

    ~~~go
    //校验数据库中是否有相同的电话号码
    	fromDB := FromDB(db)//从数据库中读取文件
    	for _, v := range fromDB {
    		//复用m
    		if m[v.Phone] == true { //这里将需要改变的标记为false
    			m[v.Phone] = false
    		}
    	}
    ~~~

    

+ 将需要创建的值和更改的值放到不同的slice中

  逐个对比从XMl文件中读入的数据与map中的标记

  ~~~go
  //将需要创建的值和更改的值放到不同的slice中
  	//这里还需要将从XML文件中读入的数据转存到数据库对应的结构体中
  	createSlice := make([]ContactsModel, 0)
  	updateSlice := make([]ContactsModel, 0)
  	for _, v := range c.C {
  		if m[v.Phone] == true {
  			createSlice = append(createSlice,
  				struct {
  					gorm.Model
  					Name  string
  					Sex   string
  					Phone string
  				}{Name: v.Name, Sex: v.Sex, Phone: v.Phone})
  			continue
  		}
  		updateSlice = append(updateSlice,
  			struct {
  				gorm.Model
  				Name  string
  				Sex   string
  				Phone string
  			}{Name: v.Name, Sex: v.Sex, Phone: v.Phone})
  	}
  ~~~

  

+ 利用gorm执行创建和修改操作

  ~~~go
  //利用gorm执行创建和修改操作
  	if len(createSlice) != 0 {
  		db.Model(&ContactsModel{}).Create(&createSlice)
  	}
  	for _, v := range updateSlice {
  		db.Model(&ContactsModel{}).Where("phone=?", v.Phone).Updates(v)
  	}
  ~~~

  

+ 在上一步中可能在执行数据库操作的某一部出错，有的数据已经执行了操作，有的没有正确执行，我们可以打开数据库事务来保证数据的一致性

  ~~~go
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
  ~~~

  如果返回的不是nil，gorm会自动执行数据库回滚操作

+ 我们执行一遍操作得到的output.xml文件中有乱码

  ~~~xml
  <Contacts>
      <contact sex="M">
          <name>&#xA;            张三&#xA;        </name>
          <phone>11111111111</phone>
      </contact>
      <contact sex="W">
          <name>&#xA;            李四&#xA;        </name>
          <phone>22222222222</phone>
      </contact>
      <contact sex="M">
          <name>&#xA;            王五&#xA;        </name>
          <phone>44444444444</phone>
      </contact>
  </Contacts>
  ~~~

  这是由于写入的时候没有处理空格和换行符引起的，这里我们添加一个函数作为处理

  ~~~go
  //去除字符串中的空格和换行符
  func replaceString(s string) (newString string) {
  	s = strings.Replace(s, "\n", "", -1)
  	newString = strings.Replace(s, " ", "", -1)
  	return
  }
  ~~~

  并将“将需要创建的值和更改的值放到不同的slice中”这一步操作修改为

  ~~~go
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
  ~~~

  



#### 代码地址

[code]: https://github.com/Tian2002/xmlDemo

