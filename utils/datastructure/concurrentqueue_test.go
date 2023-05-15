package datastructure

import (
	"errors"
	"gotest.tools/v3/assert"
	"testing"
)

func TestConcurrentQueue(t *testing.T) {
	que := NewConcurrentQueue()
	strs := [...]string{"aaa", "bbb", "ccc"}
	for _, str := range strs {
		que.Enqueue(str)
	}

	for i, _ := range strs {
		if f, _ := que.Front(); f != strs[i] {
			assert.Error(t, errors.New("queue wrong value"), "")
		}
		que.Dequeue()
	}

	if que.Empty() != true {
		assert.Error(t, errors.New("queue not empty"), "")
	}
}
