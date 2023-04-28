//  Copyright 2018 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package tasker is a task queue for the osconfig_agent.
package tasker

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/GoogleCloudPlatform/osconfig/agentconfig"
	"github.com/GoogleCloudPlatform/osconfig/clog"
)

var (
	once sync.Once
	tq = NewTaskQueue()
)

type task struct {
	run  func()
	name string
}

type TaskQueue struct {
	tc chan *task
	wg sync.WaitGroup
	mx sync.Mutex
}

func NewTaskQueue() *TaskQueue {
	q := &TaskQueue{
		tc:  make(chan *task),
	}

	return q
}

func (tq *TaskQueue) Loop(ctx context.Context) {
	tq.wg.Add(1)
	defer tq.wg.Done()

	for t := range tq.tc {
		clog.Debugf(ctx, "Tasker running %q.", t.name)
		t.run()
		clog.Debugf(ctx, "Finished task %q.", t.name)
		if agentconfig.FreeOSMemory() {
			debug.FreeOSMemory()
		}
		clog.Debugf(ctx, "Waiting for tasks to run.")
	}
}

// Enqueue adds a task to the task queue.
// Calls to Enqueue after a Close will block.
func (tq *TaskQueue) Enqueue(ctx context.Context, name string, f func()) {
	tq.mx.Lock()
	tq.tc <- &task{name: name, run: f}
	tq.mx.Unlock()
}

// Close prevents any further tasks from being enqueued and waits for the queue to empty.
// Subsequent calls to Close() will block.
func (tq *TaskQueue) Close() {
	tq.mx.Lock()
	close(tq.tc)
	tq.wg.Wait()
}


// Enqueue adds a task to the task queue.
// Calls to Enqueue after a Close will block.
func Enqueue(ctx context.Context, name string, f func()) {
	once.Do(func() {
		go tq.Loop(ctx)
	})

	tq.Enqueue(ctx, name, f)
}

// Close prevents any further tasks from being enqueued and waits for the queue to empty.
// Subsequent calls to Close() will block.
func Close() {
	tq.Close()
}
