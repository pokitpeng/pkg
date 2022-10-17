package pool

import (
	"context"
	"sync"

	"github.com/pokitpeng/pkg/concurrent/stage"
)

type Pool struct {
	index            int
	maxConcurrent    int
	mutex            sync.Mutex
	stages           []*stage.Stage
	beforeEveryStage stage.HandlerStage
	afterEveryStage  stage.HandlerStage
}

type Option func(config *Pool)

func WithMaxConcurrent(maxConcurrent int) Option {
	return func(config *Pool) {
		config.maxConcurrent = maxConcurrent
	}
}

func WithBeforeEveryStage(h stage.HandlerStage) Option {
	return func(config *Pool) {
		config.beforeEveryStage = h
	}
}

func WithAfterEveryStage(h stage.HandlerStage) Option {
	return func(config *Pool) {
		config.afterEveryStage = h
	}
}

func New(opts ...Option) *Pool {
	pool := &Pool{
		maxConcurrent: 20, // 默认最大并发数为20
	}
	for _, opt := range opts {
		opt(pool)
	}

	return pool
}

func (p *Pool) next() (int, *stage.Stage) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.stages) <= p.index {
		return 0, nil
	}
	p.index++
	return p.index, p.stages[p.index-1]
}

func (p *Pool) Run(ctx context.Context) {
	var concurrent = p.maxConcurrent
	if p.maxConcurrent >= len(p.stages) {
		concurrent = len(p.stages)
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if _, s := p.next(); s != nil {
					_ = p.run(ctx, s) // 错误已经在stage中保存
					// if err != nil && !s.ContinueWhenError {
					// 	panic(s.Desc + " error")
					// }
				} else {
					return
				}
			}
		}()
	}
	wg.Wait()
}

func (p *Pool) AddStage(desc string, h stage.Handler, opts ...stage.Option) {
	s := &stage.Stage{
		Desc:    desc,
		Handler: h,
	}
	for _, opt := range opts {
		opt(s)
	}
	p.stages = append(p.stages, s)
}

func (p *Pool) run(ctx context.Context, s *stage.Stage) (err error) {
	if s.BeforeStage != nil {
		s.BeforeStage(ctx, s)
	}
	err = s.Run(ctx)
	if p.afterEveryStage != nil {
		p.afterEveryStage(ctx, s)
	}
	return err
}
