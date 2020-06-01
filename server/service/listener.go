package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gomq/common"
	"gomq/server/consumer"
	"gomq/server/producer"
	"log"
	"net"
	"time"
)

type Listener struct {
	protocol     string
	address      string
}

func NewListener(protocol, address string) *Listener {
	return &Listener{
		protocol:     protocol,
		address:      address,
	}
}

func (l *Listener) Start(producer *producer.Receiver, consumer *consumer.Receiver) {
	listen, err := net.Listen(l.protocol, l.address)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for {
		conn, err := listen.Accept()
		fmt.Println("客户端：" + conn.RemoteAddr().String() + "连接")
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		go l.handleConn(conn, producer, consumer)
	}
}

func (l *Listener) handleConn(conn net.Conn, producerRc *producer.Receiver, consumerRc *consumer.Receiver) {
	// 如果没有可读数据，也就是读 buffer 为空，则阻塞
	ctx, cancel := context.WithCancel(context.Background())
	for {
		packet := make([]byte, 1024)
		n, err := conn.Read(packet)
		if err != nil{
			fmt.Println("客户端超时未响应或退出，关闭连接")
			_ = conn.Close()
			cancel()
			return
		}
		data := packet[:n]
		netPacket := common.Packet{}
		_ = json.Unmarshal(data, &netPacket)
		switch netPacket.Flag {
		case common.C: // 消费者连接
			_ = conn.SetDeadline(time.Now().Add(100 * time.Second))
			consumerRc.HandleConn(netPacket,conn)
			go consumerRc.Consume(ctx,conn)
		case common.CH: // 消费者心跳
			_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
			continue
		case common.P: // 生产者连接
			producerRc.Produce(netPacket.Topic,netPacket.Message)
			_ = conn.Close()
			return
		}
	}

}
