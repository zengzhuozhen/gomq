package common

import (
	"reflect"
	"testing"
)

func TestQueue_PushAndPop(t *testing.T) {
	queue := NewQueue()
	msgUnit := NewMessageUnit("A",1,Message{
		MsgKey: "aa",
		Body:   "bb",
	})

	queue.Push(msgUnit)
	queue.Push(msgUnit)
	queue.Push(msgUnit)

	retMsgUnit,err  := queue.Pop(msgUnit.Topic,0)

	if err != nil{
		t.Error(err)
	}
	if !reflect.DeepEqual(msgUnit,retMsgUnit) {
		t.Errorf("message is not equal before push and after pop")
	}
}

