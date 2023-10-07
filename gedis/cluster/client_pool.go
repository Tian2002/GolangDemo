package cluster

//gedis/cluster/cluster_pool.go
import (
	"context"
	"errors"
	"gedis/resp/client"
	pool "github.com/jolestar/go-commons-pool/v2"
)

// 实现PooledObjectFactory接口
type connectionFactory struct {
	Peer string //表示连接池连接的那个节点
}

func (f *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	cli, err := client.MakeClient(f.Peer)
	if err != nil {
		return nil, err
	}
	cli.Start() //登录客服端

	return pool.NewPooledObject(cli), nil
}

func (f *connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	c, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("类型错误")
	}
	c.Close()
	return nil
}

//下面几个方法不做具体实现

func (f *connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

func (f *connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

func (f *connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
