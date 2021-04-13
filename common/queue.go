package common

import (
	"errors"
	"sync"
)

type Queue struct {
	Local map[string][]MessageUnit // Local[Topic][Message_1,Message_2]
	mu    *sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		Local: make(map[string][]MessageUnit),
		mu:    new(sync.RWMutex),
	}
}

func (q *Queue) Push(messageUnit MessageUnit) {
	q.mu.Lock()
	// todo 并发写map问题
	q.Local[messageUnit.Topic] = append(q.Local[messageUnit.Topic], messageUnit)
	q.mu.Unlock()
}

func (q *Queue) Pop(topic string, position int64) (message MessageUnit, err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.Local[topic]) == 0 {
		return MessageUnit{}, errors.New("目前还没有消息")
	} else if int64(len(q.Local[topic])) <= position {
		return MessageUnit{}, errors.New("已消费到最新消息")
	} else {
		return q.Local[topic][position], nil
	}
}
