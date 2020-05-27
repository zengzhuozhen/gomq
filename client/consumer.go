package client

import (
	"context"
	"fmt"
	"gomq/common"
)

type Consumer struct {
	msgChan <-chan common.Message
}

func NewConsumer(msgChan chan common.Message) *Consumer {
	return &Consumer{msgChan: msgChan}
}

func (p *Consumer) Consume(ctx context.Context) common.Message {
	for {
		select {
		case msg := <-p.msgChan:
			fmt.Println("消费了:" + msg.MsgKey)
		case <-ctx.Done():
			fmt.Println("关闭消费者")
		}
	}

}
