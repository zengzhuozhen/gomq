package common

import (
	"errors"
	"sync"
)

type Queue struct {
	topic      [1024]string
	localQueue map[string] []Message
	mu sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		topic:      [1024]string{},
		localQueue: make(map[string][]Message, 1024),
	}
}

func (q *Queue) Push(topic string, message Message) {
	q.mu.Lock()
	q.localQueue[topic] = append(q.localQueue[topic],message)
	q.mu.Unlock()
}

func (q *Queue) Pop(topic string ,position int64) (message Message ,err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.localQueue[topic]) == 0{
		return Message{},errors.New("目前还没有消息")
	}else if int64(len(q.localQueue[topic]))  <= position{
		return Message{}, errors.New("已消费到最新消息")
	}else {
		return q.localQueue[topic][position] , nil
	}
}
