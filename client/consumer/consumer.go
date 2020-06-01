package consumer

import (
	"encoding/json"
	"fmt"
	"gomq/client"
	"gomq/common"
	"log"
	"time"
)

type Consumer struct {
	bc *client.BasicClient
}

func NewConsumer(protocol, host string, port, timeout int) *Consumer {
	return &Consumer{bc: &client.BasicClient{
		Protocol: protocol,
		Host:     host,
		Port:     port,
		Timeout:  timeout,
	}}
}

func (c *Consumer) Subscribe(topic string) {
	conn := c.bc.Connect()
	sendData := make([]byte, 1024)
	go func() {
		// 连接服务端
		initPosition := 0
		sendData = consumerNetPacket(topic, int64(initPosition))
		conn.Write(sendData)
		tickTimer := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-tickTimer.C:
				//fmt.Println("发送心跳包")
				conn.Write(consumeHeartPack())
			}
		}
	}()
	for {
		receData := make([]byte, 1024)
		n, err := conn.Read(receData)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(receData))
		fmt.Printf("\n")
		netPacket := common.Packet{}
		json.Unmarshal(receData[:n], &netPacket)
		fmt.Println("接收到服务端返回数据:")
	}
}

func consumerNetPacket(topic string,initPosition int64) []byte {
	netPacket := common.Packet{
		Flag:     common.C,
		Message:  common.Message{},
		Topic:    topic,
		Position: initPosition,
	}
	sendData, _ := json.Marshal(netPacket)
	return sendData
}


func consumeHeartPack() []byte {
	netPacket := common.Packet{
		Flag:          common.CH,
		Message:       common.Message{},
	}
	sendData, _ := json.Marshal(netPacket)
	return sendData
}