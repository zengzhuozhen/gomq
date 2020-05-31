package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gomq/common"
	"gomq/server/consumer"
	"gomq/server/producer"
	"io"
	"log"
	"net"
	"time"
)

type Listener struct {
	protocol string
	address  string
}

func NewListener(protocol, address string) *Listener {
	return &Listener{protocol: protocol, address: address}
}

func (l *Listener) Start(producer *producer.ProducerReceiver, consumer *consumer.ConsumerReceiver) {
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
		go handleConn(conn, producer, consumer)
	}
}

func handleConn(conn net.Conn, producer *producer.ProducerReceiver, consumer *consumer.ConsumerReceiver) {
	// 如果没有可读数据，也就是读 buffer 为空，则阻塞
	for {
		packet := make([]byte, 1024)
		n, err := conn.Read(packet)
		if err == io.EOF {	//
			fmt.Println("关闭连接")
			_ = conn.Close()
			return
		}
		data := packet[:n]
		fmt.Println(conn.RemoteAddr().String() + ": 读到的数据为" + string(data))
		netPacket := common.Packet{}
		_ = json.Unmarshal(data, &netPacket)
		switch netPacket.Flag {
		case common.C: // 消费者连接
			_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
			go consumer.Consume(context.TODO(),conn)
		case common.CH: // 消费者心跳
			_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
			continue
		case common.P: // 生产者连接
			producer.Produce(netPacket.Message)
			_ = conn.Close()
			return
		}
	}

}
