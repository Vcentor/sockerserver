// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Wei Du (duwei04@baidu.com)
// Date: 2020/8/12

package timer

import (
	"context"
	"sync"
	"time"
)

// NewOnceSuccess 创建一个只会运行成功一次的任务管理器
func NewOnceSuccess(duration time.Duration) *OnceSuccess {
	return &OnceSuccess{
		duration: duration,
	}
}

// OnceSuccess 一个只会运行成功一次的任务管理器
type OnceSuccess struct {
	duration time.Duration
	jobs     []func() error
	mu       sync.Mutex
	started  bool

	successResult map[int]bool
}

// AddJob 添加任务,在Start 之后添加任务会panic
func (os *OnceSuccess) AddJob(f func() error) {
	os.mu.Lock()
	defer os.mu.Unlock()
	if os.started {
		panic("cannot AddJob after started")
	}
	os.jobs = append(os.jobs, f)
}

// Start 开启任务，是同步的，当任务全部运行成功后，将退出
// 也可以用ctx 来控制运行的时长
func (os *OnceSuccess) Start(ctx context.Context) {
	os.mu.Lock()
	alreadyStarted := os.started
	os.started = true
	os.mu.Unlock()

	if alreadyStarted || len(os.jobs) == 0 {
		return
	}

	os.successResult = make(map[int]bool)

	tk := time.NewTicker(os.duration)
	for {
		select {
		case <-tk.C:
			os.runJobs()
			if len(os.jobs) == len(os.successResult) {
				tk.Stop()
				return
			}
		case <-ctx.Done():
			tk.Stop()
			return
		}
	}
}

func (os *OnceSuccess) runJobs() {
	for i := 0; i < len(os.jobs); i++ {
		if os.successResult[i] {
			continue
		}
		fn := os.jobs[i]
		if err := fn(); err == nil {
			os.successResult[i] = true
		}
	}
}
