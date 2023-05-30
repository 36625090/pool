/*
 * Copyright 2022 The Go Authors<36625090@qq.com>. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE file.
 */

package pool

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"io"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func runPool(total int) {

}

func runRoutine(total int) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() any {
			wg.Done()
			return fmt.Sprintf("task-%d", i)
		}()
	}
	wg.Wait()
}

//func TestNewPool(t *testing.T) {
//	run(1024)
//	//18:27:07- 18:28:25
//	//20:40:25  20:41:45
//}

func BenchmarkPool(b *testing.B) {
	b.ReportAllocs()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logfile, _ := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	logger := hclog.New(&hclog.LoggerOptions{Output: io.MultiWriter(logfile, os.Stdout), IncludeLocation: true, Level: hclog.Error})
	pool := NewPool(context.TODO(), 128, 128, 8, logger)
	total := 4096
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//var wg sync.WaitGroup
		//wg.Add(total)
		//go pool.Monitor(time.Second * 3)
		for i := 0; i < total; i++ {
			pool.SubmitTask("", func(pool *Pool, id uint64) interface{} {
				//wg.Done()
				//time.Sleep(time.Second * time.Duration(i%4))
				return fmt.Sprintf("task-%d", id)
			})
		}
	}

}

func BenchmarkRoutine(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runRoutine(4096)
	}

}

func TestPoolWait(t *testing.T) {
	total := 4096
	logfile, _ := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	logger := hclog.New(&hclog.LoggerOptions{Output: io.MultiWriter(logfile, os.Stdout), IncludeLocation: true, Level: hclog.Error})
	pool := NewPool(context.TODO(), 128, 128, 8, logger)
	for i := 0; i < total; i++ {
		pool.SubmitTask("", func(pool *Pool, id uint64) interface{} {
			//wg.Done()
			//time.Sleep(time.Second * time.Duration(i%4))
			return fmt.Sprintf("task-%d", id)
		})
	}
	time.Sleep(time.Millisecond)

	timer := time.NewTimer(time.Second * 1)
	for {
		select {
		case <-timer.C:
			t.Log("exited")
			return
		default:
			t.Log("waiting...")
		}

	}

}
