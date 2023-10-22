package stage

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStage_Run(t *testing.T) {
	// 定义一个成功的处理程序
	h1 := func(ctx context.Context) error {
		return nil
	}

	// 定义一个失败的处理程序
	h2 := func(ctx context.Context) error {
		return errors.New("failed to handle")
	}

	// 定义一个总是失败的处理程序
	h3 := func(ctx context.Context) error {
		return errors.New("always failed")
	}

	tests := []struct {
		name       string
		stage      *Stage
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "test success handler",
			stage: New(
				"test success handler",
				h1,
				WithRetryCount(3),
				WithRetryInterval(time.Second),
				WithBeforeStageFn(func(ctx context.Context) error {
					return nil
				}),
				WithAfterStageFn(func(ctx context.Context) error {
					return nil
				}),
			),
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "test failed handler without retry",
			stage: New(
				"test failed handler without retry",
				h2,
				WithRetryCount(0),
			),
			wantErr:    true,
			wantErrMsg: "failed to handle",
		},
		{
			name: "test failed handler with retry",
			stage: New(
				"test failed handler with retry",
				h2,
				WithRetryCount(3),
				WithRetryInterval(time.Second),
			),
			wantErr:    true,
			wantErrMsg: "failed to handle",
		},
		{
			name: "test always failed handler with retry",
			stage: New(
				"test always failed handler with retry",
				h3,
				WithRetryCount(3),
				WithRetryInterval(time.Second),
			),
			wantErr:    true,
			wantErrMsg: "always failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.stage.Run(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Stage.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && errors.Unwrap(err).Error() != tt.wantErrMsg {
				t.Errorf("Stage.Run() error message = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}
