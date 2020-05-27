package client

import (
	"fmt"
	"gomq/common"
)

type Producer struct {
	msgChan chan<- common.Message
}

func NewProducer(msgChan chan common.Message) *Producer {
	return &Producer{msgChan: msgChan}
}

func (p *Producer) Produce(msg common.Message) {
	fmt.Println("生产了:" + msg.MsgKey)
	p.msgChan <- msg
}
