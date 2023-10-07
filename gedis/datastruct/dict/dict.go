package dict

//gedis/datastruct/dict/dict.go

// Consumer 遍历的时候传入的方法，和sync,Map的range方法的参数类似
type Consumer func(key string, val interface{}) bool

// Dict 实现一个dict接口，当我们换底层实现时，不必修改上层的命令，比如从map+读写锁到sync.Map
type Dict interface {
	Get(key string) (val interface{}, exists bool)
	// Len 返回字典里面有多少数据
	Len() int
	Put(key string, val interface{}) (result int)
	// PutIfAbsent 如果没有才进行操作
	PutIfAbsent(key string, val interface{}) (result int)
	// PutIfExists 如果有才进行操作
	PutIfExists(key string, val interface{}) (result int)
	Remove(key string) (result int)
	ForEach(consumer Consumer)
	// Keys 返回所有键
	Keys() []string
	// RandomKeys 随机返回键
	RandomKeys(limit int) []string
	// RandomDistinctKeys 随机返回不重复的键
	RandomDistinctKeys(limit int) []string
	Clear()
}
