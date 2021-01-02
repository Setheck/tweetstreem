package app

import (
	"sync"
	"sync/atomic"
)

type History struct {
	lastIdx *int32
	history sync.Map
}

func NewHistory() *History {
	h := &History{}
	li := int32(0)
	h.lastIdx = &li
	return h
}

// Log writes the object into the history at the next index
func (h *History) Log(obj interface{}) {
	atomic.AddInt32(h.lastIdx, 1)
	h.history.Store(atomic.LoadInt32(h.lastIdx), obj)
}

// Last retrieves the object from the given index
func (h *History) Last(idx int) (interface{}, bool) {
	return h.history.Load(int32(idx))
}

func (h *History) LastIdx() int {
	return int(atomic.LoadInt32(h.lastIdx))
}

// Clear resets the history
func (h *History) Clear() {
	atomic.StoreInt32(h.lastIdx, 0)
	h.history = sync.Map{}
}
