package wait

///gedis/lib/sync/wait.go
import (
	"sync"
	"time"
)

// Wait 带超时设置的WaitGroup
type Wait struct {
	wg sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

func (w *Wait) Done() {
	w.wg.Done()
}

// WaitWithTimeout 正常Wait结束返回true，超时返回false
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		w.wg.Wait()
		ch <- struct{}{}
	}()
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return false
	}
}
