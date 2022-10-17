package pipeline

import (
	"context"

	"github.com/pokitpeng/pkg/concurrent/stage"
)

type Pipeline struct {
	stages           []*stage.Stage
	beforeEveryStage stage.HandlerStage // 对stage做一些前处理
	afterEveryStage  stage.HandlerStage // 对stage做一些后处理，一般是错误处理
	abort            bool               // 退出信号
	abortedIndex     int                // 需要退出的stages索引
	// rollback     stage.Stage        // 回滚操作
}

type Option func(p *Pipeline)

// WithBeforeEveryStage 对stage做一些前处理
func WithBeforeEveryStage(s stage.HandlerStage) Option {
	return func(config *Pipeline) {
		config.beforeEveryStage = s
	}
}

// WithAfterEveryStage 对stage做一些后处理，一般是日志操作
func WithAfterEveryStage(s stage.HandlerStage) Option {
	return func(config *Pipeline) {
		config.afterEveryStage = s
	}
}

func New(opts ...Option) *Pipeline {
	p := &Pipeline{abortedIndex: -1}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// GetStages 输出pipeline的stages
func (p *Pipeline) GetStages() []*stage.Stage {
	return p.stages
}

func (p *Pipeline) AddStage(desc string, handler stage.HandlerStage, opts ...stage.Option) {
	s := stage.Stage{
		Desc:              desc,
		Handler:           handler,
		ContinueWhenError: false,
	}
	for _, opt := range opts {
		opt(&s)
	}
	p.stages = append(p.stages, &s)
}

// AddRollback 一般用于出错终止stag后使用
// func (p *Pipeline) AddRollback(desc string, handler stage.Handler) {
// 	p.rollback = stage.Stage{
// 		Desc:    desc,
// 		Handler: handler,
// 	}
// }

func (p *Pipeline) Run(ctx context.Context) (err error) {
	for i, s := range p.stages {
		_s := s
		if err = p.run(ctx, _s); err != nil {
			p.abort = true
		}
		if p.abort && !_s.ContinueWhenError {
			p.abortedIndex = i
			break
		} else {
			p.abort = false
		}
	}
	return err
}

func (p *Pipeline) run(ctx context.Context, s *stage.Stage) (err error) {
	if s.BeforeStage != nil {
		s.BeforeStage(ctx, s)
	}
	err = s.Run(ctx)
	if p.afterEveryStage != nil {
		p.afterEveryStage(ctx, s)
	}
	return err
}
