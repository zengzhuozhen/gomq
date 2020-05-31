package producer

import (
	"fmt"
	"gomq/common"
)

type Receiver struct {
	msgChan chan<- common.Message
	queue *common.Queue
}

func NewProducerReceiver(msgChan chan common.Message, queue *common.Queue) *Receiver {
	return &Receiver{
		msgChan: msgChan,
		queue:queue,
	}
}

func (p *Receiver) Produce(topic string, msg common.Message) {
	fmt.Printf("主题 %s 生产了: %s ", topic, msg.MsgKey)
	p.queue.Push(topic,msg)
	p.msgChan<-msg
}
