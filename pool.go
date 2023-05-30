/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package pool

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	ctx            context.Context
	cancel         context.CancelFunc
	wg             *sync.WaitGroup
	taskQueue      chan *Task
	workers        []*Worker
	asyncLock      sync.RWMutex
	logger         hclog.Logger
	states         *states
	quit           chan bool
	ticker         *time.Ticker
	totalTaskCount uint64
}

func NewPool(ctx context.Context, poolQueueSize, workerCount, workerQueueSize int, logger hclog.Logger) *Pool {
	taskQueue := make(chan *Task, poolQueueSize)
	workers := make([]*Worker, workerCount)
	ctx, cancel := context.WithCancel(ctx)
	pool := &Pool{
		wg:        new(sync.WaitGroup),
		ctx:       ctx,
		cancel:    cancel,
		taskQueue: taskQueue,
		workers:   workers,
		states:    new(states),
		logger:    logger,
		ticker:    time.NewTicker(time.Millisecond * 200),
		quit:      make(chan bool),
	}

	pool.wg.Add(workerCount)

	pool.logger.Info("pool starting", "workers", workerCount, "queue", poolQueueSize)
	for i := 0; i < workerCount; i++ {
		workers[i] = &Worker{
			id:        uint64(i),
			ctx:       ctx,
			pool:      pool,
			quit:      make(chan int),
			taskQueue: make(chan *Task, workerQueueSize),
		}
		go workers[i].Start()
	}
	pool.logger.Info("pool start completed", "workers", workerCount, "queue", poolQueueSize)
	return pool
}

// SubmitTask 提交任务
// resultChan 任务的结果chan
// isFull 队列满
func (pool *Pool) SubmitTask(name string, handle TaskHandle, syncs ...bool) (resultChan chan any, isFull bool) {
	task := &Task{
		id:     atomic.AddUint64(&pool.totalTaskCount, 1),
		name:   name,
		sync:   len(syncs) > 0 && syncs[0],
		handle: handle,
		state:  TaskStateInitial,
	}
	if task.sync {
		task.resultChan = make(chan any, 1)
	}
	pool.states.Update(task.state)
	task.state = TaskStatePending

	select {
	case pool.taskQueue <- task:
	case <-pool.ticker.C:
		return nil, true
	}

	task.state = TaskStateQueued
	pool.states.Update(task.state)

	return task.resultChan, false
}

func (pool *Pool) Shutdown(force bool) {
	states := pool.states.States()
	pool.logger.Info("pool tasks statistics", "total", states.Total,
		"queued", states.Queued, "pending", states.Pending, "running",
		states.Running, "completed", states.Completed)

	pool.logger.Info("pool shutting...", "force", force)
	pool.cancel()
	close(pool.taskQueue)
	close(pool.quit)
	pool.ticker.Stop()
	pool.wg.Wait()
	pool.logger.Info("pool shutdown completed")

}

func (pool *Pool) States() *States {
	return pool.states.States()
}

func (pool *Pool) Monitor(duration time.Duration) {
	interval := time.NewTicker(duration)
	defer interval.Stop()
	for {
		select {
		case <-interval.C:
			states := pool.states.States()
			pool.logger.Info("pool statistics",
				"total", states.Total, "queued", states.Queued,
				"pending", states.Pending,
				"running", states.Running, "completed", states.Completed)
		case <-pool.quit:
			return
		}
	}
}
