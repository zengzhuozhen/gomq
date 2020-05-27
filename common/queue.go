package common

import (
	"container/list"
)

type Queue struct {
	localQueue *list.List
}

func NewQueue() *Queue {
	return &Queue{localQueue: list.New()}
}

func (q *Queue) Push(message Message) {
	 q.localQueue.PushBack(message)
}

func (q *Queue) Pop() Message {
	if ele :=  q.localQueue.Front();ele != nil{
		msg := ele.Value.(Message)
		q.localQueue.Remove(ele)
		return msg
	}
	return Message{}
}
