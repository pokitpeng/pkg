package pipeline

import (
	"context"
	"fmt"
	"github.com/pokitpeng/pkg/concurrent/stage"
	"golang.org/x/sync/semaphore"
	"sync"
)

type Pipeline struct {
	stages       []*stage.Stage
	abort        bool // 退出信号
	abortedIndex int  // 需要退出的stages索引

	async         bool // 是否异步并发执行
	maxConcurrent int
}

type Option func(p *Pipeline)

func WithMaxConcurrent(maxConcurrent int) Option {
	return func(config *Pipeline) {
		config.maxConcurrent = maxConcurrent
	}
}

func NewSync() *Pipeline {
	p := &Pipeline{abortedIndex: -1}
	return p
}

func NewAsync(opts ...Option) *Pipeline {
	p := &Pipeline{
		abortedIndex:  -1,
		async:         true,
		maxConcurrent: 10, // 默认最大10个并发
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// GetStages 输出pipeline的stages
func (p *Pipeline) GetStages() []*stage.Stage {
	return p.stages
}

func (p *Pipeline) AddStage(s *stage.Stage, opts ...stage.Option) {
	for _, opt := range opts {
		opt(s)
	}
	p.stages = append(p.stages, s)
}

func (p *Pipeline) Run(ctx context.Context) (err error) {
	if p.async {
		if err = p.asyncRun(ctx); err != nil {
			return err
		}
	} else {
		if err = p.syncRun(ctx); err != nil {
			return err
		}
	}
	return nil
}

// SyncRun 顺序同步执行
func (p *Pipeline) syncRun(ctx context.Context) (err error) {
	for i, s := range p.stages {
		_s := s
		if err = _s.Run(ctx); err != nil {
			p.abort = true
		}
		if p.abort && !_s.ContinueOnError {
			p.abortedIndex = i
			// 启动回滚流程
			for j := i; j >= 0; j-- {
				rs := p.stages[j]
				if rs.RollbackFn != nil {
					if err := rs.Rollback(ctx); err != nil {
						// error log
						fmt.Printf("%v\n", err)
					}
				}
			}
			break
		} else {
			p.abort = false
			p.abortedIndex = -1
		}
	}
	return err
}

// AsyncRun 异步并发执行
func (p *Pipeline) asyncRun(ctx context.Context) (err error) {
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(int64(p.maxConcurrent))

	// 用于处理错误和回滚
	errChan := make(chan error, len(p.stages))
	rollbackChan := make(chan int, len(p.stages))

	// 处理退出信号
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, s := range p.stages {
		_s := s

		// 等待可用的并发资源
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}

		wg.Add(1)
		go func(index int) {
			defer sem.Release(1)
			defer wg.Done()

			if runErr := _s.Run(ctx); runErr != nil {
				errChan <- runErr
				rollbackChan <- index
				cancel()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)
	close(rollbackChan)

	// 处理错误和回滚
	if len(errChan) > 0 {
		err = <-errChan
		p.abort = true
		p.abortedIndex = <-rollbackChan

		// 启动回滚流程
		for j := p.abortedIndex; j >= 0; j-- {
			rs := p.stages[j]
			if rs.RollbackFn != nil {
				if rollbackErr := rs.Rollback(ctx); rollbackErr != nil {
					// error log
					fmt.Printf("%v\n", rollbackErr)
				}
			}
		}
	} else {
		p.abort = false
		p.abortedIndex = -1
	}

	return err
}
