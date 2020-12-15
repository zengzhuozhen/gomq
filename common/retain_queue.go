package common

import "sync"

type RetainQueue struct {
	Local map[string][]MessageUnit // Local[Topic][Message_1,Message_2]
	mu    *sync.RWMutex
}

func NewRetainQueue() *RetainQueue {
	local := make(map[string][]MessageUnit,1024)
	return &RetainQueue{Local: local, mu: new(sync.RWMutex)}
}

func (q *RetainQueue) Push(messageUnit MessageUnit) {
	q.mu.RLock()
	q.Local[messageUnit.Topic] = append(q.Local[messageUnit.Topic], messageUnit)
	q.mu.RUnlock()
}

func (q *RetainQueue) ReadAll(topic string) []MessageUnit {
	return q.Local[topic]
}
