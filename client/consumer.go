package client

import (
	"encoding/json"
	"fmt"
	"gomq/common"
	"log"
	"time"
)

type Consumer struct {
	bc *BasicClient
}

func NewConsumer(protocol, host string, port, timeout int) *Consumer {
	return &Consumer{bc: &BasicClient{
		protocol: protocol,
		host:     host,
		port:     port,
		timeout:  timeout,
	}}
}

func (c *Consumer) Subscribe() {
	conn := c.bc.Connect()
	sendData := make([]byte, 1024)
	go func() {
		// 连接服务端
		sendData = consumerNetPacket(common.C)
		conn.Write(sendData)
		tickTimer := time.NewTicker(1 * time.Second)
		for {
			sendData = consumerNetPacket(common.CH)
			select {
			case <-tickTimer.C:
				fmt.Println("发送心跳包")
				conn.Write(sendData)
			}
		}
	}()
	for {
		receData := make([]byte, 1024)
		n, err := conn.Read(receData)
		if err != nil {
			log.Fatal(err)
		}
		netPacket := common.Packet{}
		json.Unmarshal(receData[:n], &netPacket)
		fmt.Println("接收到服务端返回数据:")
		fmt.Printf("%+v", netPacket)
	}
}

func consumerNetPacket(class common.Flag) []byte {
	netPacket := common.NewPacket(class, common.Message{})
	sendData, _ := json.Marshal(netPacket)
	return sendData
}
