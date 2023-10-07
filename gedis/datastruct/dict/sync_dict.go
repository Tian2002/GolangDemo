package dict

//gedis/datastruct/dict/sync_dict.go
import "sync"

type SyncDict struct {
	m sync.Map
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
	return dict.m.Load(key)
}

func (dict *SyncDict) Len() int {
	l := 0
	dict.m.Range(func(key, value any) bool {
		l++
		return true
	})
	return l
}

func (dict *SyncDict) Put(key string, val interface{}) (result int) {
	_, exists := dict.m.Load(key)
	dict.m.Store(key, val)
	if exists { //原本存在key返回0，而不是失败了
		return 0
	}
	return 1
}

// PutIfAbsent 没有这个key才操作
func (dict *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, exists := dict.m.Load(key)
	if !exists {
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

// PutIfExists 存在这个key才操作
func (dict *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, exists := dict.m.Load(key)
	if exists {
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

func (dict *SyncDict) Remove(key string) (result int) {
	_, exists := dict.m.Load(key)
	if exists {
		dict.m.Delete(key)
		return 1
	}
	return 0
}

func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.m.Range(func(key, value any) bool {
		consumer(key.(string), value)
		return true //这里一直返回true，让他施加到所有的k，v上
	})
}

func (dict *SyncDict) Keys() []string {
	res := make([]string, dict.Len())
	i := 0
	dict.m.Range(func(key, value any) bool {
		res[i] = key.(string)
		i++
		return true
	})
	return res
}

func (dict *SyncDict) RandomKeys(limit int) []string {
	res := make([]string, limit)
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value any) bool {
			res[i] = key.(string)
			return false //每次去随机取一个
		})
	}
	return res
}

func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	res := make([]string, limit)
	i := 0
	//一次随机取limit个
	dict.m.Range(func(key, value any) bool {
		res[i] = key.(string)
		i++
		return i < limit
	})
	return res
}

func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict()
}
