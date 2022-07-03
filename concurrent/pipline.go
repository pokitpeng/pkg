package concurrent

// import (
// 	"sync/atomic"
// 	"time"

// 	"git.ucloudadmin.com/ubase/ubase/context"
// 	"golang.org/x/sync/errgroup"
// )

// type stageHandle func(ctx context.Context) error

// type Pipeline struct {
// 	stages   []stageHandle
// 	rollback stageHandle
// 	async    bool
// }

// func NewPipeline(async bool) *Pipeline {
// 	return &Pipeline{async: async}
// }

// func (p *Pipeline) AddStage(h stageHandle) {
// 	p.stages = append(p.stages, h)
// }

// func (p *Pipeline) AddRollback(h stageHandle) {
// 	p.rollback = h
// }

// func (p *Pipeline) syncRun(ctx context.Context) error {
// 	for _, stage := range p.stages {
// 		if aborted(ctx) {
// 			// ctx.Info("request aborted")
// 			return nil
// 		}
// 		atomic.AddInt64(&_task, 1)
// 		if err := stage(ctx); err != nil {
// 			atomic.AddInt64(&_task, -1)
// 			return err
// 		}
// 		atomic.AddInt64(&_task, -1)
// 	}
// 	return nil
// }

// func (p *Pipeline) asyncRun(ctx context.Context) error {
// 	e := errgroup.Group{}
// 	for index := range p.stages {
// 		_index := index
// 		e.Go(func() error {
// 			atomic.AddInt64(&_task, 1)
// 			defer atomic.AddInt64(&_task, -1)
// 			return p.stages[_index](ctx)
// 		})
// 	}
// 	return e.Wait()
// }

// func (p *Pipeline) Run(ctx context.Context) (err error) {
// 	if p.async {
// 		err = p.asyncRun(ctx)
// 	} else {
// 		err = p.syncRun(ctx)
// 	}
// 	if err != nil && p.rollback != nil {
// 		// _ctx := context.Named("rollback_" + context.GetSession(ctx))
// 		if _err := p.rollback(_ctx); _err != nil {
// 			//todo 告警
// 			// _ctx.Error("rollback error", "err", _err.Error())
// 		}
// 	}
// 	return err
// }

// func Abort(ctx context.Context) error {
// 	ctx.Set("done", true)
// 	return nil
// }

// func aborted(ctx context.Context) bool {
// 	return ctx.GetBool("done")
// }

// var _task int64

// func TaskCount() (uint64, string) {
// 	if _task <= 0 {
// 		return 0, ""
// 	}
// 	return uint64(_task), ""
// }

// func WaitTaskEmpty() {
// 	for {
// 		if _task != 0 {
// 			time.Sleep(time.Second)
// 			continue
// 		}
// 		break
// 	}
// }
