package event

import (
	"sync"
	"sync/atomic"
)

// Event 将来可能发生的一次性事件。
type Event struct {
	fired int32
	c     chan struct{}
	o     sync.Once
}

// Fire 触发结束事件，多次并发调用是安全的。只有第一次对Fire的调用导致通道关闭，返回true。
func (e *Event) Fire() bool {
	ret := false
	e.o.Do(func() {
		atomic.StoreInt32(&e.fired, 1)
		close(e.c)
		ret = true
	})
	return ret
}

// Done 返回一个通道，该通道将在调用Fire时关闭。
func (e *Event) Done() <-chan struct{} {
	return e.c
}

// HasFired 如果Fire被调用，返回true。
func (e *Event) HasFired() bool {
	return atomic.LoadInt32(&e.fired) == 1
}

// NewEvent 返回一个新的、随时可用的Event。
func NewEvent() *Event {
	return &Event{c: make(chan struct{})}
}
