package concurrent

import (
	"context"
	"fmt"
	"runtime/debug"

	"golang.org/x/sync/errgroup"
)

type handler func(ctx context.Context) error
type handlerStage func(ctx context.Context, s *stage)

type stage struct {
	desc              string
	handler           handler
	error             error // 保存处理过程中出现的错误
	continueWhenError bool
}

type StageOption func(s *stage)

// WithContinueWhenError ...
func WithContinueWhenError(b bool) StageOption {
	return func(config *stage) {
		config.continueWhenError = b
	}
}

func (s *stage) run(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s occur panic: %v\n", s.desc, string(debug.Stack()))
		}
		s.error = err
	}()
	err = s.handler(ctx)
	return err
}

type Pipeline struct {
	stages       []*stage
	rollback     stage
	beforeStage  handlerStage // 对stage做一些前处理
	afterStage   handlerStage // 对stage做一些后处理，一般是错误处理
	async        bool
	abort        bool
	abortedIndex int
}

type PipelineOption func(p *Pipeline)

// WithBeforeStage 对stage做一些前处理
func WithBeforeStage(s handlerStage) PipelineOption {
	return func(config *Pipeline) {
		config.beforeStage = s
	}
}

// WithAfterStage 对stage做一些后处理，一般是错误处理
func WithAfterStage(s handlerStage) PipelineOption {
	return func(config *Pipeline) {
		config.afterStage = s
	}
}

func newPipeline(async bool) *Pipeline {
	return &Pipeline{async: async, abortedIndex: -1}
}

func NewSyncPipeline(opts ...PipelineOption) *Pipeline {
	p := newPipeline(false)
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func NewAsyncPipeline(opts ...PipelineOption) *Pipeline {
	p := newPipeline(true)
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// GetStages 输出pipeline的stages
func (p *Pipeline) GetStages() []*stage {
	return p.stages
}

// HandleStagesErrors 处理stages中出现的错误
func (p *Pipeline) HandleStagesErrors(handlerError func(...interface{})) {
	for i, s := range p.stages { // 输出处理过程中所有的error
		if s.error != nil {
			handlerError(s.desc, s.error)
			// fmt.Printf("%s occur error: %v\n", s.desc, s.error)
		}
		if i == p.abortedIndex {
			break
		}
	}
}

func (p *Pipeline) AddStage(desc string, handler handler, opts ...StageOption) {
	s := stage{
		desc:              desc,
		handler:           handler,
		continueWhenError: false,
	}
	for _, opt := range opts {
		opt(&s)
	}
	p.stages = append(p.stages, &s)
}

// AddRollback 一般用于出错终止stag后使用
func (p *Pipeline) AddRollback(desc string, handler handler) {
	p.rollback = stage{
		desc:    desc,
		handler: handler,
	}
}

func (p *Pipeline) syncRun(ctx context.Context) error {
	var err error
	for i, s := range p.stages {
		_s := s
		if p.beforeStage != nil {
			p.beforeStage(ctx, _s)
		}
		if err = _s.run(ctx); err != nil {
			p.abort = true
		}
		if p.afterStage != nil {
			p.afterStage(ctx, _s)
		}
		if p.abort && !_s.continueWhenError {
			p.abortedIndex = i
			// fmt.Printf("pipeline aborted\n")
			break
		} else {
			p.abort = false
		}
	}
	return err
}

func (p *Pipeline) asyncRun(ctx context.Context) error {
	e := errgroup.Group{}
	for _, s := range p.stages {
		_s := s
		e.Go(func() error {
			if p.beforeStage != nil {
				p.beforeStage(ctx, _s)
			}
			err := _s.run(ctx)
			if p.afterStage != nil {
				p.afterStage(ctx, _s)
			}
			return err
		})
	}
	// 只会抛出第一个出错stage的error
	return e.Wait()
}

func (p *Pipeline) Run(ctx context.Context) (err error) {
	if p.async {
		err = p.asyncRun(ctx)
	} else {
		err = p.syncRun(ctx)
	}
	if err != nil && p.rollback.handler != nil {
		if e := p.rollback.handler(ctx); e != nil {
			if p.afterStage != nil {
				p.afterStage(ctx, &p.rollback)
			}
		}
	}
	return err
}

// Go 对原生的goroutine添加recover保护
func Go(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("panic happened: %v\n", string(debug.Stack()))
			}
		}()
		f()
	}()
}
