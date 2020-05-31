package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"gomq/common"
	"log"
	"net"
)

type Receiver struct {
	msgChan <-chan common.Message
	pool    *Pool
}

func NewConsumerReceiver(msgChan chan common.Message, pool *Pool) *Receiver {
	return &Receiver{msgChan: msgChan, pool: pool}
}

func (c *Receiver) Consume(ctx context.Context, conn net.Conn) {
	for {
		select {
		case msg := <-c.msgChan:
			connUid := conn.RemoteAddr().String()
			fmt.Println("ip:" +connUid)
			netPacket := common.Packet{
				Flag:          common.S,
				Message:       msg,
				Topic:         c.pool.Topic[connUid],
			}
			if data, err := json.Marshal(netPacket); err == nil {
				fmt.Printf("准备发送给客户端消费者: %s %+v", conn.RemoteAddr(), netPacket)
				conn.Write(data)
			} else {
				log.Fatalln(err.Error())
			}
		case <-ctx.Done():
			fmt.Println("消费者连接已关闭，退出消费循环")
			c.pool.State[conn.RemoteAddr().String()] = false
			return
		}
	}
}


