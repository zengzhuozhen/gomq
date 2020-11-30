package common

import (
	"errors"
	"sync"
)

type Queue struct {
	Local          map[string][]MessageUnit // Local[Topic][Message_1,Message_2]
	offset         map[string]int64         // offset[Topic]1024
	mu             *sync.RWMutex
	openPersistent bool
	PersistentChan chan MessageUnit
}

func NewQueue(openPersistent bool) *Queue {
	return &Queue{
		Local:          make(map[string][]MessageUnit, 1024),
		offset:         make(map[string]int64),
		mu:             new(sync.RWMutex),
		openPersistent: openPersistent,
		PersistentChan: make(chan MessageUnit),
	}
}

func (q *Queue) Push(messageUnit MessageUnit) {
	q.mu.RLock()
	if len(q.Local[messageUnit.Topic]) >= 1024 { // 超过 1024个元素 清空一下
		q.offset[messageUnit.Topic] += int64(len(q.Local[messageUnit.Topic]))
		q.Local[messageUnit.Topic] = make([]MessageUnit, 0)
	}
	q.Local[messageUnit.Topic] = append(q.Local[messageUnit.Topic], messageUnit)
	q.mu.RUnlock()
	if q.openPersistent{
		q.PersistentChan <- messageUnit
	}

}

func (q *Queue) Pop(topic string, position int64) (message MessageUnit, err error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	position -= q.offset[topic]
	// todo
	// 新加进来的消费者会从第一个开始消费，此时 position < offset , 需要实现类似缺页中断的功能，
	// 并且把之后所有的message都放在一个临时buffer里，(局部性原理)
	if len(q.Local[topic]) == 0 {
		return MessageUnit{}, errors.New("目前还没有消息")
	} else if int64(len(q.Local[topic])) <= position {
		return MessageUnit{}, errors.New("已消费到最新消息")
	} else {
		return q.Local[topic][position], nil
	}
}