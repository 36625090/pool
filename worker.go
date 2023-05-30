/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package pool

import (
	"context"
	"sync"
	"time"
)

type Worker struct {
	sync.RWMutex
	ctx       context.Context
	id        uint64
	pool      *Pool
	taskQueue chan *Task
	quit      chan int
}

func (worker *Worker) Start() {
	go func() {
		for {
			select {
			case task, ok := <-worker.pool.taskQueue:
				if !ok {
					worker.pool.logger.Warn("receive task with a closed chan", "worker", worker.id)
					close(worker.quit)
					return
				}
				worker.taskQueue <- task
				worker.pool.logger.Trace("pool queue -> worker queue",
					"worker", worker.id, "length", len(worker.pool.taskQueue), "task", task.id, "length", len(worker.taskQueue))
			case <-worker.quit:
				return
			}
		}
	}()
	run := true
	for run {
		select {
		case task, ok := <-worker.taskQueue:
			if !ok {
				worker.pool.logger.Warn("receive task with a closed chan", "worker", worker.id)
				run = false
			}
			now := time.Now()
			worker.pool.logger.Trace("worker started", "worker", worker.id, "task", task.id, "name", task.name)
			task.state = TaskStateRunning
			worker.pool.states.Update(task.state)

			result := task.handle(worker.pool, task.id)
			task.state = TaskStateCompleted

			worker.pool.states.Update(task.state)

			worker.pool.logger.Trace("work executed", "worker", worker.id, "task", task.id, "name", task.name, "latency", time.Now().Sub(now))

			if task.sync {
				task.resultChan <- result
			}

		case <-worker.ctx.Done():
			close(worker.quit)
		case <-worker.quit:
			run = false
		}
	}
	worker.pool.wg.Done()
	worker.pool.logger.Trace("worker exited", "worker", worker.id, "queued", len(worker.taskQueue))
}
