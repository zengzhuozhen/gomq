package producer

import (
	"fmt"
	"gomq/common"
)

type ProducerReceiver struct {
	msgChan chan<- common.Message
}

func NewProducerReceiver(msgChan chan common.Message) *ProducerReceiver {
	return &ProducerReceiver{msgChan: msgChan}
}

func (p *ProducerReceiver) Produce(msg common.Message) {
	fmt.Println("生产了:" + msg.MsgKey)
	p.msgChan <- msg
}
