package pipeline

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pokitpeng/pkg/concurrent/stage"
)

var stage1 = func(ctx context.Context, s *stage.Stage) error {
	fmt.Println("i am stage 1")
	return nil
}
var stage2 = func(ctx context.Context, s *stage.Stage) error {
	fmt.Println("i am stage 2")
	return nil
}

var stage3Panic = func(ctx context.Context, s *stage.Stage) error {
	fmt.Println("i am stage 3")
	panic("occur panic")
	return nil
}

var stage3Error = func(ctx context.Context, s *stage.Stage) error {
	fmt.Println("i am stage 3")
	return errors.New("some error")
}

var stage4 = func(ctx context.Context, s *stage.Stage) error {
	time.Sleep(time.Millisecond * 200)
	fmt.Println("i am stage 4")
	return errors.New("some error")
}

var rollback = func(ctx context.Context, s *stage.Stage) error {
	fmt.Println("rollback")
	return nil
}

func TestNewPipeline_Sync_Normal(t *testing.T) {
	ctx := context.Background()
	pipeline := New()
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Println(err)
	}
}

func TestNewPipeline_Sync_Error(t *testing.T) {
	ctx := context.Background()
	pipeline := New(
		WithAfterEveryStage(func(ctx context.Context, s *stage.Stage) error {
			if s.Error != nil {
				fmt.Printf("alarm ==> %v error: %v\n", s.Desc, s.Error)
			}
			return nil
		}))
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	pipeline.AddStage("run stage 3", stage3Error, stage.WithContinueWhenError(true))
	pipeline.AddStage("run stage 4", stage4)
	// pipeline.AddRollback("rollback", rollback)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}

func TestNewPipeline_Sync_Panic(t *testing.T) {
	ctx := context.Background()
	pipeline := New(
		WithAfterEveryStage(func(ctx context.Context, s *stage.Stage) error {
			if s.Error != nil {
				fmt.Printf("alarm ==> %v error: %v\n", s.Desc, s.Error)
			}
			return nil
		}))
	pipeline.AddStage("run stage 1", stage1)
	pipeline.AddStage("run stage 2", stage2)
	pipeline.AddStage("run stage 3", stage3Panic, stage.WithContinueWhenError(true))
	pipeline.AddStage("run stage 4", stage4)
	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("pipeline run error: %v\n", err) // 此处只能输出最后一个出错stage的error
	}
}
