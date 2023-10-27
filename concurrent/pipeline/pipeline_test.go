package pipeline

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pokitpeng/pkg/concurrent/stage"
)

var stage1 = stage.New(
	"stage1",
	func(ctx context.Context) error {
		fmt.Println("i am stage 1")
		return nil
	},
)

var stage1Rollback = func(ctx context.Context) error {
	fmt.Println("i am stage 1 rollback")
	return nil
}

var stage2 = stage.New(
	"stage2",
	func(ctx context.Context) error {
		fmt.Println("i am stage 2")
		return nil
	},
)

var stage2Rollback = func(ctx context.Context) error {
	fmt.Println("i am stage 2 rollback")
	return nil
}

var stage3Panic = stage.New(
	"stage3",
	func(ctx context.Context) error {
		fmt.Println("i am stage 3")
		panic("occur panic")
		return nil
	},
)

var stage3Error = stage.New(
	"stage3",
	func(ctx context.Context) error {
		fmt.Println("i am stage 3")
		return errors.New("some error")
	},
)

var stage4 = stage.New(
	"stage4",
	func(ctx context.Context) error {
		fmt.Println("i am stage 4")
		return nil
	},
)

func TestNewPipeline_Sync_Normal(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSync()
	pipeline.AddStage(stage1)
	pipeline.AddStage(stage2)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Println(err)
	}
}

func TestNewPipeline_Error(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSync()
	pipeline.AddStage(stage1)
	pipeline.AddStage(stage2)
	pipeline.AddStage(stage3Error, stage.WithRetryCount(3)) // 出错后重试三次
	pipeline.AddStage(stage4)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
		fmt.Printf("row error: %v\n", errors.Unwrap(err))
	}
}

func TestNewPipeline_Error_Continue(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSync()
	pipeline.AddStage(stage1)
	pipeline.AddStage(stage2)
	pipeline.AddStage(stage3Error, stage.WithRetryCount(3), stage.WithContinueOnError()) // 出错了还继续往后运行
	pipeline.AddStage(stage4)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个stage的error
	}
}

func TestNewPipeline_Panic(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSync()
	pipeline.AddStage(stage1)
	pipeline.AddStage(stage2)
	pipeline.AddStage(stage3Panic, stage.WithContinueOnError()) // 出错依旧会继续
	pipeline.AddStage(stage3Error)                              // 此处出错会结束pipeline
	pipeline.AddStage(stage4)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}

func TestNewPipeline_Rollback(t *testing.T) {
	ctx := context.Background()
	pipeline := NewSync()
	pipeline.AddStage(stage1, stage.WithRollbackFn(stage1Rollback))
	pipeline.AddStage(stage2, stage.WithRollbackFn(stage2Rollback))
	pipeline.AddStage(stage3Error) // 此处出错会依次向上找回滚函数回滚，结束pipeline
	pipeline.AddStage(stage4)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}

func TestPool_Normal(t *testing.T) {
	ctx := context.Background()
	pipeline := NewAsync(
		WithMaxConcurrent(5),
	)

	for i := 0; i < 10; i++ {
		index := i
		pipeline.AddStage(stage.New(
			fmt.Sprintf("%d", index),
			func(ctx context.Context) error {
				fmt.Printf("run stage %d\n", index)
				time.Sleep(time.Second)
				return nil
			}),
		)
	}
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}

func TestPool_Error(t *testing.T) {
	ctx := context.Background()
	pipeline := NewAsync(
		WithMaxConcurrent(5),
	)

	for i := 0; i < 20; i++ {
		index := i
		pipeline.AddStage(stage.New(
			fmt.Sprintf("%d", index),
			func(ctx context.Context) error {
				if index == 22 {
					return fmt.Errorf("%d exec error", index)
				}
				fmt.Printf("run stage %d\n", index)
				time.Sleep(time.Second)
				return nil
			}),
		)
	}
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}
