package datastructure

import "sync"

type IConcurrentQueue interface {
	Enqueue(item interface{})
	Dequeue() (item interface{}, exist bool)
	Front() (interface{}, bool)
	Empty() bool
	GetContent() []interface{}
	SetContent([]interface{})
	Length() int
}

func NewConcurrentQueue() IConcurrentQueue {
	return &ConcurrentQueue{}
}

type ConcurrentQueue struct {
	queue []interface{}
	mu    sync.Mutex
}

func (q *ConcurrentQueue) GetContent() []interface{} {
	return q.queue
}

func (q *ConcurrentQueue) SetContent(qu []interface{}) {
	q.queue = qu
}

func (q *ConcurrentQueue) Enqueue(item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, item)
}

func (q *ConcurrentQueue) Dequeue() (interface{}, bool) {
	if q.Empty() {
		return nil, false
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	temp := q.queue[0]
	q.queue = q.queue[1:]
	return temp, true
}

func (q *ConcurrentQueue) Front() (interface{}, bool) {
	if q.Empty() {
		return nil, false
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.queue[0], true
}

func (q *ConcurrentQueue) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue) == 0
}

func (q *ConcurrentQueue) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}
