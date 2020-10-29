package common

import (
	"errors"
	"sync"
)

type Queue struct {
	local           map[string][]Message // local[Topic][Message_1,Message_2]
	offset          map[string]int64     // offset[Topic]1024
	mu              *sync.RWMutex
	PersistentChan  chan MessageUnit
	MembersSyncChan chan MessageUnit
}

func NewQueue() *Queue {
	return &Queue{
		local:           make(map[string][]Message, 1024),
		offset:          make(map[string]int64),
		mu:              new(sync.RWMutex),
		PersistentChan:  make(chan MessageUnit),
		MembersSyncChan: make(chan MessageUnit),
	}
}

func (q *Queue) Push(topic string, message Message) {
	q.mu.RLock()
	if len(q.local[topic]) >= 1024 { // 超过 1024个元素 清空一下
		q.offset[topic] += int64(len(q.local[topic]))
		q.local[topic] = make([]Message,0)
	}
	q.local[topic] = append(q.local[topic], message)
	q.mu.RUnlock()
	messageUnit := NewMessageUnit(topic, message)
	q.PersistentChan <- messageUnit
	q.MembersSyncChan <- messageUnit
}

func (q *Queue) Pop(topic string, position int64) (message Message, err error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	position -= q.offset[topic]
	// todo
	// 新加进来的消费者会从第一个开始消费，此时 position < offset , 需要实现类似缺页中断的功能，
	// 并且把之后所有的message都放在一个临时buffer里，(局部性原理)
	if len(q.local[topic]) == 0 {
		return Message{}, errors.New("目前还没有消息")
	} else if int64(len(q.local[topic])) <= position {
		return Message{}, errors.New("已消费到最新消息")
	} else {
		return q.local[topic][position], nil
	}
}
