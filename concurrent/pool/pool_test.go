package pool

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pokitpeng/pkg/concurrent/stage"
)

func TestPool_Normal(t *testing.T) {
	pool := New(
		WithMaxConcurrent(5),
		WithAfterEveryStage(func(c context.Context, s *stage.Stage) {
			if s.Error != nil {
				fmt.Printf("%s error ==> %v\n", s.Desc, s.Error)
				return
			}
			fmt.Printf("%s success\n", s.Desc)
		}),
	)

	for i := 0; i < 50; i++ {
		index := i
		pool.AddStage(
			fmt.Sprintf("index %v", i),
			func(ctx context.Context) error {
				time.Sleep(time.Millisecond * 500)
				fmt.Println(fmt.Sprintf("i am worker %d", index))
				return nil
			},
		)
	}
	pool.Run(context.Background())
	fmt.Println("all task done !!!")
}

func TestPool_Panic(t *testing.T) {
	pool := New(
		WithMaxConcurrent(5),
		WithAfterEveryStage(func(c context.Context, s *stage.Stage) {
			if s.Error != nil {
				fmt.Printf("%s error ==> %v\n", s.Desc, s.Error)
				return
			}
			fmt.Printf("%s success\n", s.Desc)
		}),
	)

	for i := 0; i < 50; i++ {
		index := i
		if index == 20 {
			pool.AddStage(
				fmt.Sprintf("index %v", i),
				func(ctx context.Context) error {
					time.Sleep(time.Millisecond * 500)
					panic("mock panic")
					fmt.Println(fmt.Sprintf("i am worker %d", index))
					return nil
				},
			)
			continue
		}
		pool.AddStage(
			fmt.Sprintf("index %v", i),
			func(ctx context.Context) error {
				time.Sleep(time.Millisecond * 500)
				fmt.Println(fmt.Sprintf("i am worker %d", index))
				return nil
			},
		)
	}
	pool.Run(context.Background())
	fmt.Println("all task done !!!")
}
