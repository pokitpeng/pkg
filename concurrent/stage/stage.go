package stage

import (
	"context"
	"fmt"
	"runtime/debug"
)

type Handler func(ctx context.Context) error
type HandlerStage func(ctx context.Context, s *Stage) error

type Stage struct {
	Desc              string       // 描述信息
	BeforeStage       HandlerStage // 前处理
	Handler           HandlerStage // 实际工作函数
	AfterStage        HandlerStage // 后处理
	Error             error        // 保存处理过程中出现的错误
	ContinueWhenError bool         // 出错后是否继续
}

type Option func(s *Stage)

// WithContinueWhenError ...
func WithContinueWhenError(b bool) Option {
	return func(config *Stage) {
		config.ContinueWhenError = b
	}
}

// WithBeforeStage 对Stage做一些前处理
func WithBeforeStage(s HandlerStage) Option {
	return func(config *Stage) {
		config.BeforeStage = s
	}
}

// WithAfterStage 对Stage做一些后处理，一般是错误处理
func WithAfterStage(s HandlerStage) Option {
	return func(config *Stage) {
		config.AfterStage = s
	}
}

func (s *Stage) Run(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s occur panic: %v\n", s.Desc, string(debug.Stack()))
		}
		s.Error = err
	}()
	err = s.Handler(ctx, s)
	return err
}
