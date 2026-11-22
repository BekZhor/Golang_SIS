package concurrency

import (
	"fmt"
	"sort"
	"sync"
)

const (
	QueueCapacity = 5
	WorkerCount   = 3
)

type QueueItem struct {
	TaskID    int
	Operation int
}

type TaskQueue struct {
	mu      sync.Mutex
	jobs    chan QueueItem
	items   map[int]QueueItem
	changed chan struct{}
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		jobs:    make(chan QueueItem, QueueCapacity),
		items:   make(map[int]QueueItem),
		changed: make(chan struct{}),
	}
}

func (q *TaskQueue) signal() {
	close(q.changed)
	q.changed = make(chan struct{})
}

func (q *TaskQueue) Enqueue(item QueueItem) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.items[item.TaskID]; exists {
		return fmt.Errorf("task %d is already in queue", item.TaskID)
	}

	select {
	case q.jobs <- item:
		q.items[item.TaskID] = item
		q.signal()
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

func (q *TaskQueue) Dequeue(stopCh <-chan struct{}) (QueueItem, bool) {
	for {
		q.mu.Lock()

		select {
		case <-stopCh:
			q.mu.Unlock()
			return QueueItem{}, false
		default:
		}

		select {
		case item := <-q.jobs:
			delete(q.items, item.TaskID)
			q.mu.Unlock()
			return item, true
		default:
		}

		changed := q.changed
		q.mu.Unlock()

		select {
		case <-changed:
		case <-stopCh:
			return QueueItem{}, false
		}
	}
}

func (q *TaskQueue) Get(id int) (QueueItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, exists := q.items[id]
	return item, exists
}

func (q *TaskQueue) List() []QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make([]QueueItem, 0, len(q.items))

	for _, item := range q.items {
		result = append(result, item)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TaskID < result[j].TaskID
	})

	return result
}

func (q *TaskQueue) Remove(id int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.items[id]; !exists {
		return fmt.Errorf("task %d not found in queue", id)
	}

	count := len(q.jobs)

	for i := 0; i < count; i++ {
		item := <-q.jobs

		if item.TaskID != id {
			q.jobs <- item
		}
	}

	delete(q.items, id)
	q.signal()

	return nil
}

func (q *TaskQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *TaskQueue) Capacity() int {
	return cap(q.jobs)
}

type TaskHandler func(QueueItem) error

type ErrorHandler func(QueueItem, error)

type Worker struct {
	id      int
	queue   *TaskQueue
	handler TaskHandler
	onError ErrorHandler
	stopCh  <-chan struct{}
}

func NewWorker(
	id int,
	queue *TaskQueue,
	handler TaskHandler,
	onError ErrorHandler,
	stopCh <-chan struct{},
) *Worker {
	return &Worker{
		id:      id,
		queue:   queue,
		handler: handler,
		onError: onError,
		stopCh:  stopCh,
	}
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		item, ok := w.queue.Dequeue(w.stopCh)
		if !ok {
			return
		}

		fmt.Printf("Worker %d picked up task %d\n", w.id, item.TaskID)

		err := w.handler(item)

		if err != nil {
			if w.onError != nil {
				w.onError(item, err)
			}
		}
	}
}

type WorkerPool struct {
	queue   *TaskQueue
	handler TaskHandler
	onError ErrorHandler

	wg sync.WaitGroup
	mu sync.Mutex

	running  bool
	stopping bool
	stopCh   chan struct{}
}

func NewWorkerPool(
	handler TaskHandler,
	onError ErrorHandler,
) *WorkerPool {
	return &WorkerPool{
		queue:   NewTaskQueue(),
		handler: handler,
		onError: onError,
	}
}

func (p *WorkerPool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("worker pool is already running")
	}

	if p.handler == nil {
		return fmt.Errorf("task handler is nil")
	}

	if p.stopping {
		return fmt.Errorf("worker pool is stopping")
	}

	p.stopCh = make(chan struct{})
	p.running = true

	for i := 1; i <= WorkerCount; i++ {
		worker := NewWorker(
			i,
			p.queue,
			p.handler,
			p.onError,
			p.stopCh,
		)

		p.wg.Add(1)
		go worker.Start(&p.wg)
	}

	fmt.Println("Worker Pool started")
	return nil
}

func (p *WorkerPool) Submit(item QueueItem) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopping {
		return fmt.Errorf("worker pool is stopping")
	}

	return p.queue.Enqueue(item)
}

func (p *WorkerPool) Stop() {
	p.mu.Lock()

	if !p.running {
		p.mu.Unlock()
		fmt.Println("Worker Pool is not running")
		return
	}

	p.stopping = true
	p.running = false

	close(p.stopCh)

	p.mu.Unlock()

	p.wg.Wait()

	p.mu.Lock()
	p.stopping = false
	p.mu.Unlock()

	fmt.Println("Worker Pool stopped")
}

func (p *WorkerPool) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.running
}

func (p *WorkerPool) ListQueued() []QueueItem {
	return p.queue.List()
}

func (p *WorkerPool) GetQueued(id int) (QueueItem, bool) {
	return p.queue.Get(id)
}

func (p *WorkerPool) RemoveQueued(id int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopping {
		return fmt.Errorf("worker pool is stopping")
	}

	return p.queue.Remove(id)
}
