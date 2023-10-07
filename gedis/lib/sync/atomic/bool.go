package atomic

///gedis/lib/sync/atomic.go
import "sync/atomic"

// Boolean 原子性的bool类型
type Boolean uint32

func (b *Boolean) Get() bool {
	return atomic.LoadUint32((*uint32)(b)) != 0
}

func (b *Boolean) Set(val bool) {
	if val {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
