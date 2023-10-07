package resp

type Connection interface {
	Write([]byte) error //向客户端写
	GetDBIndex() int    //获取现在使用的是那个db
	SelectDB(index int) //切换数据库
}
