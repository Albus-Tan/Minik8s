package datastructure

import (
	"fmt"
	"strings"
)

// IQueue interface that all queues implement
type IQueue interface {
	Enqueue(value interface{})
	Dequeue() (value interface{}, ok bool)
	Peek() (value interface{}, ok bool)

	Empty() bool
	Size() int
	Clear()
	Values() []interface{}
	String() string
}

// Assert Queue implementation
var _ IQueue = (*Queue)(nil)

// Queue holds elements in a singly-linked-list
type Queue struct {
	list *List
}

// NewQueue instantiates a new empty queue
func NewQueue() *Queue {
	return &Queue{list: NewList()}
}

// Enqueue adds a value to the end of the queue
func (queue *Queue) Enqueue(value interface{}) {
	queue.list.Add(value)
}

// Dequeue removes first element of the queue and returns it, or nil if queue is empty.
// Second return parameter is true, unless the queue was empty and there was nothing to dequeue.
func (queue *Queue) Dequeue() (value interface{}, ok bool) {
	value, ok = queue.list.Get(0)
	if ok {
		queue.list.Remove(0)
	}
	return
}

// Peek returns first element of the queue without removing it, or nil if queue is empty.
// Second return parameter is true, unless the queue was empty and there was nothing to peek.
func (queue *Queue) Peek() (value interface{}, ok bool) {
	return queue.list.Get(0)
}

// Empty returns true if queue does not contain any elements.
func (queue *Queue) Empty() bool {
	return queue.list.Empty()
}

// Size returns number of elements within the queue.
func (queue *Queue) Size() int {
	return queue.list.Size()
}

// Clear removes all elements from the queue.
func (queue *Queue) Clear() {
	queue.list.Clear()
}

// Values returns all elements in the queue (FIFO order).
func (queue *Queue) Values() []interface{} {
	return queue.list.Values()
}

// String returns a string representation of container
func (queue *Queue) String() string {
	str := "LinkedListQueue\n"
	values := []string{}
	for _, value := range queue.list.Values() {
		values = append(values, fmt.Sprintf("%v", value))
	}
	str += strings.Join(values, ", ")
	return str
}

// Check that the index is within bounds of the list
func (queue *Queue) withinRange(index int) bool {
	return index >= 0 && index < queue.list.Size()
}
