package config

import (
	"bufio"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type config struct {
	TypeInt   int      `cfg:"TypeInt"`
	TypeStr   string   `cfg:"TypeString"`
	TypeBool  bool     `cfg:"TypeBool"`
	TypeSlice []string `cfg:"TypeSlice"`
}

// 默认配置，为了测试就随便写入几个值,通常默认配置不会全部都有默认配置，只会有几个必须项，有的是需要配置开启后才会有接下来的其他配置
var defaultProperties *config = &config{
	TypeInt:   -1,
	TypeStr:   "defaultProperties",
	TypeBool:  true,
	TypeSlice: []string{"-1", "0", "1"},
}

// Properties 全局变量，可以让所有人拿到的实际配置
var Properties *config = new(config)

func parse(src io.Reader) *config {
	scanner := bufio.NewScanner(src)
	m := make(map[string]string) //记录每一对配置的key，val
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) > 0 && text[0] == '#' || len(text) == 0 { //改行是注释
			continue
		}
		index := strings.Index(text, " ")
		if index > 0 && index < len(text)-1 { //合法的定义
			key, val := strings.ToLower(text[0:index]), text[index+1:] //将key转化为小写是为了大小写不敏感
			m[key] = val
		}
	}
	if err := scanner.Err(); err != nil { //这里错误返回的是非EOF的错误，所以不用自己再判断
		panic(err)
	}

	/*
		之前都是文件读取相关操作，下面是反射相关的内容
		1:一个个去除结构体里面tag标签的值
		2：判断前面存的map里面有没有这个标签值
		3：有的话就看这个元素定义的类型，将数据写进去
	*/
	typeOf := reflect.TypeOf(Properties)
	valueOf := reflect.ValueOf(Properties)
	//1:一个个去除结构体里面tag标签的值
	for i := 0; i < typeOf.Elem().NumField(); i++ {
		fieldType := typeOf.Elem().Field(i)
		//可能这个元素没有tag：cfg标签
		tagName, ok := fieldType.Tag.Lookup("cfg")
		if ok {
			//2：判断前面存的map里面有没有这个标签值
			tagValue, exists := m[strings.ToLower(tagName)] //将key转化为小写是为了做大小写不敏感的匹配
			//3：有的话就看这里对应的类型，将数据写进去
			if exists {
				switch typeOf.Elem().Field(i).Type.Kind() {
				case reflect.String:
					valueOf.Elem().Field(i).SetString(tagValue)
				case reflect.Int, reflect.Int32, reflect.Int64:
					parseInt, err := strconv.ParseInt(tagValue, 10, 64)
					if err != nil {
						continue
					}
					valueOf.Elem().Field(i).SetInt(parseInt)
				case reflect.Bool:
					valueOf.Elem().Field(i).SetBool(tagValue == "true") //这里直接写，而不判断配置文件里面值是否符合规范，是因为不符合规范这里就会取默认值，还是false
				case reflect.Slice:
					sli := strings.Split(tagValue, ",")
					valueOf.Elem().Field(i).Set(reflect.ValueOf(sli))
				}
			}
		}
	}

	return Properties
}

func SetUpConfig(filePath string) {
	if filePath == "" { //使用默认配置
		Properties = defaultProperties
		return
	}
	//判断文件配置是否存在以及是否是文件夹而非文件
	info, err := os.Stat(filePath)
	//配置文件配置错误直接panic
	if err != nil {
		panic(err)
		return
	}
	if info.IsDir() {
		panic("配置文件路径错误，需要一个文件而不是文件夹")
		return
	}
	if err == nil && !info.IsDir() {
		src, err := os.Open(filePath)
		if err != nil {
			panic(err)
			return
		}
		defer src.Close()
		parse(src)
	}
}
