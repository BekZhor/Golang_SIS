package concurrency

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testTimeout = 3 * time.Second

func receiveTask(t *testing.T, ch <-chan int) int {
	t.Helper()

	select {
	case id := <-ch:
		return id
	case <-time.After(testTimeout):
		t.Fatal("timeout waiting for task")
		return 0
	}
}

// -------------------- TaskQueue --------------------

func TestNewTaskQueue(t *testing.T) {
	q := NewTaskQueue()

	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got %d items", q.Len())
	}

	if q.Capacity() != QueueCapacity {
		t.Fatalf("expected capacity %d, got %d", QueueCapacity, q.Capacity())
	}
}

func TestEnqueueAndDequeue(t *testing.T) {
	q := NewTaskQueue()

	item := QueueItem{TaskID: 1, Operation: 2}

	if err := q.Enqueue(item); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	got, exists := q.Get(1)
	if !exists {
		t.Fatal("task not found")
	}

	if got != item {
		t.Fatalf("expected %+v, got %+v", item, got)
	}

	stopCh := make(chan struct{})

	got, ok := q.Dequeue(stopCh)
	if !ok {
		t.Fatal("dequeue failed")
	}

	if got != item {
		t.Fatalf("expected %+v, got %+v", item, got)
	}

	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got %d items", q.Len())
	}
}

func TestDuplicateTaskID(t *testing.T) {
	q := NewTaskQueue()

	first := QueueItem{TaskID: 1, Operation: 1}
	second := QueueItem{TaskID: 1, Operation: 2}

	if err := q.Enqueue(first); err != nil {
		t.Fatal(err)
	}

	if err := q.Enqueue(second); err == nil {
		t.Fatal("expected duplicate task error")
	}

	if q.Len() != 1 {
		t.Fatalf("expected 1 item, got %d", q.Len())
	}
}

func TestQueueCapacity(t *testing.T) {
	q := NewTaskQueue()

	for i := 1; i <= QueueCapacity; i++ {
		err := q.Enqueue(QueueItem{TaskID: i, Operation: 1})
		if err != nil {
			t.Fatalf("enqueue task %d failed: %v", i, err)
		}
	}

	err := q.Enqueue(QueueItem{TaskID: QueueCapacity + 1, Operation: 1})
	if err == nil {
		t.Fatal("expected queue full error")
	}

	if q.Len() != QueueCapacity {
		t.Fatalf("expected %d items, got %d", QueueCapacity, q.Len())
	}
}

func TestQueueList(t *testing.T) {
	q := NewTaskQueue()

	ids := []int{3, 1, 2}

	for _, id := range ids {
		if err := q.Enqueue(QueueItem{TaskID: id, Operation: 1}); err != nil {
			t.Fatal(err)
		}
	}

	items := q.List()

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	for i, expectedID := range []int{1, 2, 3} {
		if items[i].TaskID != expectedID {
			t.Fatalf("expected task %d at index %d, got %d",
				expectedID, i, items[i].TaskID)
		}
	}
}

func TestQueueRemove(t *testing.T) {
	q := NewTaskQueue()

	for i := 1; i <= 3; i++ {
		if err := q.Enqueue(QueueItem{TaskID: i, Operation: 1}); err != nil {
			t.Fatal(err)
		}
	}

	if err := q.Remove(2); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	if _, exists := q.Get(2); exists {
		t.Fatal("removed task still exists")
	}

	if q.Len() != 2 {
		t.Fatalf("expected 2 items, got %d", q.Len())
	}

	if err := q.Remove(100); err == nil {
		t.Fatal("expected error when removing nonexistent task")
	}
}

func TestDequeueStops(t *testing.T) {
	q := NewTaskQueue()
	stopCh := make(chan struct{})

	close(stopCh)

	_, ok := q.Dequeue(stopCh)
	if ok {
		t.Fatal("expected dequeue to stop")
	}
}

// -------------------- WorkerPool --------------------

func TestWorkerPoolStartup(t *testing.T) {
	pool := NewWorkerPool(func(QueueItem) error {
		return nil
	}, nil)

	if pool.IsRunning() {
		t.Fatal("pool should not be running initially")
	}

	if err := pool.Start(); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	if !pool.IsRunning() {
		t.Fatal("pool should be running")
	}

	if err := pool.Start(); err == nil {
		t.Fatal("expected error when starting an already running pool")
	}

	pool.Stop()

	if pool.IsRunning() {
		t.Fatal("pool should not be running after Stop")
	}
}

func TestWorkerPoolRejectsNilHandler(t *testing.T) {
	pool := NewWorkerPool(nil, nil)

	if err := pool.Start(); err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestAsynchronousSubmission(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	pool := NewWorkerPool(func(item QueueItem) error {
		close(started)
		<-release
		return nil
	}, nil)

	if err := pool.Start(); err != nil {
		t.Fatal(err)
	}

	if err := pool.Submit(QueueItem{TaskID: 1, Operation: 1}); err != nil {
		t.Fatal(err)
	}

	// Submit returns even though the handler is still running.
	select {
	case <-started:
	case <-time.After(testTimeout):
		t.Fatal("worker did not start processing")
	}

	close(release)
	pool.Stop()
}

func TestMultipleTasks(t *testing.T) {
	processed := make(chan int, 5)

	pool := NewWorkerPool(func(item QueueItem) error {
		processed <- item.TaskID
		return nil
	}, nil)

	if err := pool.Start(); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 5; i++ {
		if err := pool.Submit(QueueItem{TaskID: i, Operation: 1}); err != nil {
			t.Fatal(err)
		}
	}

	received := make(map[int]bool)

	for i := 0; i < 5; i++ {
		id := receiveTask(t, processed)

		if received[id] {
			t.Fatalf("task %d processed more than once", id)
		}

		received[id] = true
	}

	pool.Stop()

	if len(received) != 5 {
		t.Fatalf("expected 5 processed tasks, got %d", len(received))
	}
}

func TestConcurrentTaskProcessing(t *testing.T) {
	var active atomic.Int32
	var maxActive atomic.Int32

	started := make(chan struct{}, WorkerCount)
	release := make(chan struct{})

	pool := NewWorkerPool(func(item QueueItem) error {
		current := active.Add(1)
		defer active.Add(-1)

		for {
			old := maxActive.Load()
			if current <= old || maxActive.CompareAndSwap(old, current) {
				break
			}
		}

		started <- struct{}{}
		<-release

		return nil
	}, nil)

	if err := pool.Start(); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= WorkerCount; i++ {
		if err := pool.Submit(QueueItem{TaskID: i, Operation: 1}); err != nil {
			t.Fatal(err)
		}
	}

	// Wait until at least two workers are executing simultaneously.
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(testTimeout):
			close(release)
			pool.Stop()
			t.Fatal("tasks did not start concurrently")
		}
	}

	if maxActive.Load() < 2 {
		close(release)
		pool.Stop()
		t.Fatalf("expected at least 2 concurrent workers, got %d", maxActive.Load())
	}

	close(release)
	pool.Stop()
}

func TestWorkerPoolStopAndRestart(t *testing.T) {
	processed := make(chan int, 2)

	pool := NewWorkerPool(func(item QueueItem) error {
		processed <- item.TaskID
		return nil
	}, nil)

	if err := pool.Start(); err != nil {
		t.Fatal(err)
	}

	if err := pool.Submit(QueueItem{TaskID: 1, Operation: 1}); err != nil {
		t.Fatal(err)
	}

	if id := receiveTask(t, processed); id != 1 {
		t.Fatalf("expected task 1, got %d", id)
	}

	pool.Stop()

	if pool.IsRunning() {
		t.Fatal("pool should be stopped")
	}

	// Submissions are allowed while the pool is stopped.
	if err := pool.Submit(QueueItem{TaskID: 2, Operation: 1}); err != nil {
		t.Fatalf("submit while stopped failed: %v", err)
	}

	if len(pool.ListQueued()) != 1 {
		t.Fatal("task should remain in queue")
	}

	if err := pool.Start(); err != nil {
		t.Fatalf("restart failed: %v", err)
	}

	if id := receiveTask(t, processed); id != 2 {
		t.Fatalf("expected task 2, got %d", id)
	}

	pool.Stop()
}

func TestStopWaitsForActiveTasks(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	pool := NewWorkerPool(func(item QueueItem) error {
		close(started)
		<-release
		return nil
	}, nil)

	if err := pool.Start(); err != nil {
		t.Fatal(err)
	}

	if err := pool.Submit(QueueItem{TaskID: 1, Operation: 1}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(testTimeout):
		close(release)
		pool.Stop()
		t.Fatal("task did not start")
	}

	stopped := make(chan struct{})

	go func() {
		pool.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		close(release)
		t.Fatal("Stop returned before active task finished")
	case <-time.After(50 * time.Millisecond):
		// Stop is waiting for the active handler.
	}

	close(release)

	select {
	case <-stopped:
	case <-time.After(testTimeout):
		t.Fatal("Stop did not return after task completion")
	}
}

func TestStopPreservesQueuedTasks(t *testing.T) {
	started := make(chan struct{}, QueueCapacity)
	release := make(chan struct{})

	pool := NewWorkerPool(func(item QueueItem) error {
		started <- struct{}{}
		<-release
		return nil
	}, nil)

	if err := pool.Start(); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= QueueCapacity; i++ {
		if err := pool.Submit(QueueItem{
			TaskID:    i,
			Operation: 1,
		}); err != nil {
			t.Fatal(err)
		}
	}

	// Wait until all three workers have picked up tasks.
	for i := 0; i < WorkerCount; i++ {
		select {
		case <-started:
		case <-time.After(testTimeout):
			close(release)
			pool.Stop()
			t.Fatal("not all workers started")
		}
	}

	// The remaining tasks should still be in the queue.
	beforeStop := len(pool.ListQueued())

	if beforeStop != QueueCapacity-WorkerCount {
		close(release)
		pool.Stop()
		t.Fatalf("expected %d queued tasks before Stop, got %d",
			QueueCapacity-WorkerCount, beforeStop)
	}

	stopped := make(chan struct{})

	go func() {
		pool.Stop()
		close(stopped)
	}()

	close(release)

	select {
	case <-stopped:
	case <-time.After(testTimeout):
		t.Fatal("Stop did not finish")
	}

	afterStop := pool.ListQueued()

	if len(afterStop) != beforeStop {
		t.Fatalf("expected %d queued tasks after Stop, got %d",
			beforeStop, len(afterStop))
	}
}

func TestErrorHandler(t *testing.T) {
	expectedErr := errors.New("processing failed")
	errorsReceived := make(chan error, 1)

	pool := NewWorkerPool(
		func(item QueueItem) error {
			return expectedErr
		},
		func(item QueueItem, err error) {
			errorsReceived <- err
		},
	)

	if err := pool.Start(); err != nil {
		t.Fatal(err)
	}

	if err := pool.Submit(QueueItem{TaskID: 1, Operation: 1}); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errorsReceived:
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected %v, got %v", expectedErr, err)
		}
	case <-time.After(testTimeout):
		t.Fatal("error handler was not called")
	}

	pool.Stop()
}

// -------------------- Concurrent access --------------------

func TestConcurrentQueueAccess(t *testing.T) {
	q := NewTaskQueue()

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			_ = q.Enqueue(QueueItem{
				TaskID:    id,
				Operation: 1,
			})

			q.List()
			q.Len()
			q.Capacity()
			q.Get(id)
		}(i)
	}

	wg.Wait()

	if q.Len() > q.Capacity() {
		t.Fatalf("queue length %d exceeds capacity %d", q.Len(), q.Capacity())
	}
}
