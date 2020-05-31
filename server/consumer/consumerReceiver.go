package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"gomq/common"
	"log"
	"net"
)

type ConsumerReceiver struct {
	msgChan    <-chan common.Message
	Ctx        context.Context
}

func NewConsumerReceiver(msgChan chan common.Message) *ConsumerReceiver {
	return &ConsumerReceiver{msgChan: msgChan}
}

func (c *ConsumerReceiver) Consume(ctx context.Context,conn net.Conn) {
	for {
		select {
		case msg := <-c.msgChan:
			netPacket := common.NewPacket(common.S, msg)
			if data, err := json.Marshal(netPacket); err == nil {
				fmt.Printf("准备发送给客户端消费者: %s %+v",conn.RemoteAddr(), netPacket )
				conn.Write(data)
			} else {
				log.Fatalln(err.Error())
			}
		case <-ctx.Done():
			fmt.Println("关闭消费者连接")
			conn.Close()
			return
		}
	}
}
