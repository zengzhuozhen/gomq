package common

import (
	"errors"
	"sync"
)

// Queue 无需保存信息，因为只保存retain为0的消息，但是需要给各个客户端分发消息，他们的消费速度不一样，所以需要保存到map中
// 但是随着请求的增加，MessageUnit队列的大小会增加，由于MessageUnit队列的大小和Pool里保存的Offset相关，不能随意删掉，所以需要解决这个问题
// todo
type Queue struct {
	Local map[string][]MessageUnit // Local[Topic][Message_1,Message_2]
	Mutex *sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		Local: make(map[string][]MessageUnit),
		Mutex: new(sync.Mutex),
	}
}

func (q *Queue) Push(messageUnit MessageUnit) {
	q.Mutex.Lock()
	q.Local[messageUnit.Topic] = append(q.Local[messageUnit.Topic], messageUnit)
	q.Mutex.Unlock()
}

func (q *Queue) Pop(topic string, position int64) (message MessageUnit, err error) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	if len(q.Local[topic]) == 0 {
		return MessageUnit{}, errors.New("目前还没有消息")
	} else if int64(len(q.Local[topic])) <= position {
		return MessageUnit{}, errors.New("已消费到最新消息")
	} else {
		return q.Local[topic][position], nil
	}
}

func (q *Queue)GetLen(topic string) int {
	return len(q.Local[topic])
}