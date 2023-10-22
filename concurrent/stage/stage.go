package stage

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

type Handler func(ctx context.Context) error

type Stage struct {
	Desc    string  // 描述信息
	Handler Handler // 实际工作函数
	Error   error   // 保存处理过程中出现的错误

	// 可选参数
	ContinueOnError bool          // 出错后继续运行，for pipeline
	BeforeStageFn   Handler       // 前处理
	AfterStageFn    Handler       // 后处理
	RollbackFn      Handler       // 出错后回滚
	RetryCount      int32         // 出错后重试次数
	RetryInterval   time.Duration // 出错后重试间隔
}

type Option func(s *Stage)

// WithContinueOnError 出错后继续运行，for pipeline
func WithContinueOnError() Option {
	return func(config *Stage) {
		config.ContinueOnError = true
	}
}

// WithBeforeStageFn 对Stage做一些前处理
func WithBeforeStageFn(s Handler) Option {
	return func(config *Stage) {
		config.BeforeStageFn = s
	}
}

// WithAfterStageFn 对Stage做一些后处理，一般是错误处理
func WithAfterStageFn(s Handler) Option {
	return func(config *Stage) {
		config.AfterStageFn = s
	}
}

func WithRollbackFn(s Handler) Option {
	return func(Stage *Stage) {
		Stage.RollbackFn = s
	}
}

func WithRetryCount(c int32) Option {
	return func(Stage *Stage) {
		Stage.RetryCount = c
	}
}

func WithRetryInterval(c time.Duration) Option {
	return func(Stage *Stage) {
		Stage.RetryInterval = c
	}
}

func New(desc string, handler Handler, opts ...Option) *Stage {
	s := &Stage{
		Desc:    desc,
		Handler: handler,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Stage) Run(ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s occur panic: %v\n", s.Desc, string(debug.Stack()))
		}
		s.Error = err
	}()

	// Before stage
	if s.BeforeStageFn != nil {
		if err = s.BeforeStageFn(ctx); err != nil {
			return fmt.Errorf("before stage error: %w", err)
		}
	}

	if s.RetryCount <= 0 {
		s.RetryCount = 0
	}
	// Retry loop
	for retry := 0; retry <= int(s.RetryCount); retry++ {
		if err = s.Handler(ctx); err != nil {
			if retry == int(s.RetryCount) {
				return fmt.Errorf("run handler[last retry count:%d] error: %w", retry, err)
			}
		}
		if s.RetryInterval > 0 {
			time.Sleep(s.RetryInterval)
		}
	}

	// After stage
	if s.AfterStageFn != nil {
		if err = s.AfterStageFn(ctx); err != nil {
			return fmt.Errorf("after stage error: %w", err)
		}
	}

	//是否回滚由上层决定，默认不自动触发
	//if err != nil && s.RollbackFn != nil {
	//	if err = s.RollbackFn(ctx); err != nil {
	//		return fmt.Errorf("RollbackFn error: %w", err)
	//	}
	//}
	return err
}

func (s *Stage) Rollback(ctx context.Context) (err error) {
	err = s.RollbackFn(ctx)
	if err != nil {
		return fmt.Errorf("%s RollbackFn error: %w", s.Desc, err)
	}
	return err
}
