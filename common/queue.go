package common

import (
	"errors"
	"sync"
)

type Queue struct {
	Local          map[string][]MessageUnit // Local[Topic][Message_1,Message_2]
	mu             *sync.RWMutex
	openPersistent bool
	PersistentChan chan MessageUnit
}

func NewQueue(openPersistent bool) *Queue {
	return &Queue{
		Local:          make(map[string][]MessageUnit, 1024),
		mu:             new(sync.RWMutex),
		openPersistent: openPersistent,
		PersistentChan: make(chan MessageUnit),
	}
}

func (q *Queue) Push(messageUnit MessageUnit) {
	q.mu.RLock()
	q.Local[messageUnit.Topic] = append(q.Local[messageUnit.Topic], messageUnit)
	q.mu.RUnlock()
	if q.openPersistent{
		q.PersistentChan <- messageUnit
	}
}

func (q *Queue) Pop(topic string, position int64) (message MessageUnit, err error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.Local[topic]) == 0 {
		return MessageUnit{}, errors.New("目前还没有消息")
	} else if int64(len(q.Local[topic])) <= position {
		return MessageUnit{}, errors.New("已消费到最新消息")
	} else {
		return q.Local[topic][position], nil
	}
}
