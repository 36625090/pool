/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package pool

import "sync"

type TaskState int

const (
	TaskStateUnknown           = -2
	TaskStateFailure           = -1
	TaskStateInitial TaskState = iota
	TaskStatePending
	TaskStateQueued
	TaskStateRunning
	TaskStateCompleted
)

func (m TaskState) String() string {
	switch m {
	case TaskStateFailure:
		return "failure"
	case TaskStateInitial:
		return "initial"
	case TaskStateQueued:
		return "queued"
	case TaskStatePending:
		return "pending"
	case TaskStateRunning:
		return "running"
	case TaskStateCompleted:
		return "completed"
	}
	return "unknown"
}

type States struct {
	Total     uint64
	Queued    uint64
	Pending   uint64
	Running   uint64
	Completed uint64
}

type states struct {
	sync.RWMutex
	total     uint64
	queued    uint64
	pending   uint64
	running   uint64
	completed uint64
}

func (m *states) Update(state TaskState) {
	m.Lock()
	switch state {
	case TaskStateFailure:
	case TaskStateInitial:
		fallthrough
	case TaskStatePending:
		m.total += 1
		m.pending += 1
	case TaskStateQueued:
		m.pending -= 1
		m.queued += 1
	case TaskStateRunning:
		m.queued -= 1
		m.running += 1
	case TaskStateCompleted:
		m.running -= 1
		m.completed += 1
	}
	m.Unlock()
}
func (m *states) Queued() uint64 {
	var n uint64 = 0
	m.Lock()
	n = m.queued
	m.Unlock()
	return n
}
func (m *states) Pending() uint64 {
	var n uint64 = 0
	m.Lock()
	n = m.pending
	m.Unlock()
	return n
}
func (m *states) Running() uint64 {
	var n uint64 = 0
	m.Lock()
	n = m.running
	m.Unlock()
	return n
}
func (m *states) Completed() uint64 {
	var n uint64 = 0
	m.Lock()
	n = m.completed
	m.Unlock()
	return n
}
func (m *states) States() *States {
	var state States
	m.Lock()
	state.Total = m.total
	state.Running = m.running
	state.Completed = m.completed
	state.Pending = m.pending
	state.Queued = m.queued
	m.Unlock()
	return &state
}
