/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package pool

type TaskHandle func(pool *Pool, id uint64) any

type Task struct {
	id         uint64
	name       string
	sync       bool
	handle     TaskHandle
	state      TaskState
	resultChan chan interface{}
}
