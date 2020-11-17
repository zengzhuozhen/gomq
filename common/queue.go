package common

import (
	"errors"
	"sync"
)

type Queue struct {
	local           map[string][]MessageUnit // local[Topic][Message_1,Message_2]
	offset          map[string]int64         // offset[Topic]1024
	mu              *sync.RWMutex
	PersistentChan  chan MessageUnit
	MembersSyncChan chan MessageUnit
}

func NewQueue() *Queue {
	return &Queue{
		local:           make(map[string][]MessageUnit, 1024),
		offset:          make(map[string]int64),
		mu:              new(sync.RWMutex),
		PersistentChan:  make(chan MessageUnit),
		MembersSyncChan: make(chan MessageUnit),
	}
}

func (q *Queue) Push(messageUnit MessageUnit, needPersist, needSync bool) {
	q.mu.RLock()
	if len(q.local[messageUnit.Topic]) >= 1024 { // 超过 1024个元素 清空一下
		q.offset[messageUnit.Topic] += int64(len(q.local[messageUnit.Topic]))
		q.local[messageUnit.Topic] = make([]MessageUnit, 0)
	}
	q.local[messageUnit.Topic] = append(q.local[messageUnit.Topic], messageUnit)
	q.mu.RUnlock()
	if needPersist {
		q.PersistentChan <- messageUnit
	}
	if needSync {
		q.MembersSyncChan <- messageUnit
	}
}

func (q *Queue) Pop(topic string, position int64) (message MessageUnit, err error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	position -= q.offset[topic]
	// todo
	// 新加进来的消费者会从第一个开始消费，此时 position < offset , 需要实现类似缺页中断的功能，
	// 并且把之后所有的message都放在一个临时buffer里，(局部性原理)
	if len(q.local[topic]) == 0 {
		return MessageUnit{}, errors.New("目前还没有消息")
	} else if int64(len(q.local[topic])) <= position {
		return MessageUnit{}, errors.New("已消费到最新消息")
	} else {
		return q.local[topic][position], nil
	}
}
