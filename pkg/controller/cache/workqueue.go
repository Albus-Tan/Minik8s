package cache

import (
	"minik8s/utils/datastructure"
	"sync"
)

type WorkQueue interface {
	datastructure.IQueue

	// Close the queue
	Close()
}

func NewWorkQueue() WorkQueue {
	w := &workQueue{
		queue:  datastructure.NewQueue(),
		closed: false,
	}
	w.cond.L = &w.lock
	return w
}

type workQueue struct {
	queue datastructure.IQueue
	// processing set
	lock sync.Mutex
	cond sync.Cond

	// Indication the queue is closed.
	// Used to indicate a queue is closed so a control loop can exit when a queue is empty.
	// Currently, not used to gate any of CRUD operations.
	closed bool
}

func (w *workQueue) Close() {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.closed = true
	w.cond.Broadcast()
}

func (w *workQueue) Enqueue(value interface{}) {
	w.cond.L.Lock()
	w.queue.Enqueue(value)
	w.cond.L.Unlock()
	w.cond.Signal()
}

func (w *workQueue) Dequeue() (value interface{}, ok bool) {
	w.cond.L.Lock()
	if w.queue.Empty() {
		w.cond.Wait()
	}
	value, ok = w.queue.Dequeue()
	w.cond.L.Unlock()
	return value, ok
}

func (w *workQueue) Peek() (value interface{}, ok bool) {
	w.cond.L.Lock()
	value, ok = w.queue.Peek()
	w.cond.L.Unlock()
	return value, ok
}

func (w *workQueue) Empty() bool {
	w.cond.L.Lock()
	isEmpty := w.queue.Empty()
	w.cond.L.Unlock()
	return isEmpty
}

func (w *workQueue) Size() int {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	return w.queue.Size()
}

func (w *workQueue) Clear() {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.queue.Clear()
}

func (w *workQueue) Values() []interface{} {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	return w.queue.Values()
}

func (w *workQueue) String() string {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	return w.queue.String()
}
