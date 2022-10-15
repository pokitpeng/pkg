package concurrent

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

var stage1 = func(ctx context.Context) error {
	fmt.Println("i am stage 1")
	return nil
}
var stage2 = func(ctx context.Context) error {
	fmt.Println("i am stage 2")
	return nil
}

var stage3Panic = func(ctx context.Context) error {
	fmt.Println("i am stage 3")
	panic("occur panic")
	return nil
}

var stage3Error = func(ctx context.Context) error {
	fmt.Println("i am stage 3")
	return errors.New("some error")
}

var stage4 = func(ctx context.Context) error {
	time.Sleep(time.Millisecond * 200)
	fmt.Println("i am stage 4")
	return errors.New("some error")
}

var rollback = func(ctx context.Context) error {
	fmt.Println("rollback")
	return nil
}

func TestNewPipeline_Sync_Normal(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSyncPipeline()
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Println(err)
	}
}

func TestNewPipeline_Sync_Error(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSyncPipeline(WithAfterStage(func(ctx context.Context, s *stage) {
		if s.error != nil {
			fmt.Printf("alarm ==> %v error: %v\n", s.desc, s.error)
		}
	}))
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	pipeline.AddStage("run stage 3", stage3Error, WithContinueWhenError(true))
	pipeline.AddStage("run stage 4", stage4)
	pipeline.AddRollback("rollback", rollback)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}

func TestNewPipeline_Sync_Panic(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSyncPipeline(WithAfterStage(func(ctx context.Context, s *stage) {
		if s.error != nil {
			fmt.Printf("alarm ==> %v error: %v\n", s.desc, s.error)
		}
	}))
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	pipeline.AddStage("run stage 3", stage3Panic, WithContinueWhenError(true))
	pipeline.AddStage("run stage 4", stage4)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}

func TestNewPipeline_Async_Normal(t *testing.T) {
	ctx := context.Background()
	pipeline := NewAsyncPipeline()
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err)
	}
}

func TestNewPipeline_Async_Error(t *testing.T) {
	ctx := context.Background()
	pipeline := NewAsyncPipeline(
		WithBeforeStage(func(ctx context.Context, s *stage) {
			s.desc += " change desc name"
		}),
		WithAfterStage(func(ctx context.Context, s *stage) {
			if s.error != nil {
				fmt.Printf("alarm ==> %v error: %v\n", s.desc, s.error)
			}
		}),
	)
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	pipeline.AddStage("run stage 3", stage3Error)
	pipeline.AddStage("run stage 4", stage4)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出第一个出错stage的error
	}
}

func TestNewPipeline_Async_Panic(t *testing.T) {
	ctx := context.Background()
	pipeline := NewAsyncPipeline(WithAfterStage(func(ctx context.Context, s *stage) {
		if s.error != nil {
			fmt.Printf("alarm ==> %v error: %v\n", s.desc, s.error)
		}
	}))
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	pipeline.AddStage("run stage 3", stage3Panic)
	pipeline.AddStage("run stage 4", stage4)
	if err := pipeline.Run(ctx); err != nil {
		// fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出第一个出错stage的error
	}
}
