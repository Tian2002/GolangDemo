package consistenthash

//gedis/lib/consistenthash/consistenthash.go

import (
	"hash/crc32"
	"sort"
)

type HashFunc func(data []byte) uint32

type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int          //这里不使用uint32是因为后面会进行排序，我们想要使用go里面自带的排序函数
	nodeHashMap map[int]string //nodeHash->节点
}

func NewNodeMap(fn HashFunc) *NodeMap {
	if fn == nil {
		fn = crc32.ChecksumIEEE //默认的方法
	}
	return &NodeMap{
		hashFunc:    fn,
		nodeHashMap: make(map[int]string),
	}
}

// IsEmpty 判断是否初始化
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

// AddNode 传入某个节点的标识，将他放入节点中。
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodeHashMap[hash] = key
	}
	sort.Ints(m.nodeHashs)
}

// PickNode 判断每个key取那个节点
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	index := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	if index == len(m.nodeHashs) {
		index = 0
	}
	return m.nodeHashMap[m.nodeHashs[index]]
}
