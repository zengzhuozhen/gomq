package common

import (
	"errors"
	"fmt"
	"sync"
)

type Queue struct {
	topic           [1024]string
	localQueue      map[string][]Message
	offset          map[string]int64
	mu              *sync.RWMutex
	PersistentChan  chan MessageUnit
	MembersSyncChan chan MessageUnit
}

func NewQueue() *Queue {
	return &Queue{
		topic:           [1024]string{},
		localQueue:      make(map[string][]Message, 1024),
		offset:          make(map[string]int64),
		mu:              new(sync.RWMutex),
		PersistentChan:  make(chan MessageUnit),
		MembersSyncChan: make(chan MessageUnit),
	}
}

func (q *Queue) Push(topic string, message Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	fmt.Println(topic)
	if len(q.localQueue[topic]) >= 1024 { // 超过 1024个元素 清空一下
		q.offset[topic] += int64(len(q.localQueue[topic]))
		q.localQueue[topic] = make([]Message, 1024)
	}
	q.localQueue[topic] = append(q.localQueue[topic], message)
	messageUnit := NewPersistentUnit(topic, message)
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
	if len(q.localQueue[topic]) == 0 {
		return Message{}, errors.New("目前还没有消息")
	} else
	if int64(len(q.localQueue[topic])) <= position {
		return Message{}, errors.New("已消费到最新消息")
	} else
	{
		return q.localQueue[topic][position], nil
	}
}

func (q *Queue) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()

}
