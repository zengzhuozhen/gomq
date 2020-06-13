package consumer

import (
	"encoding/json"
	"fmt"
	"gomq/client"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"net"
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

func (c *Consumer) Subscribe(topic []string) {
	conn, err := c.bc.Connect()
	if err != nil {
		panic("连接服务端失败")
	}
	go func() {
		// 连接服务端
		//initPosition := 0
		//sendData = consumerNetPacket(topic, int64(initPosition))
		subscribePacket := protocolPacket.NewSubscribePacket(1, topic, 0)
		err := subscribePacket.Write(conn)
		if err != nil {
			fmt.Println("客户端订阅主题失败:发送subscribe")
		}
		var fh protocolPacket.FixedHeader
		if err := fh.Read(conn); err != nil {
			fmt.Errorf("读取包头失败", err)
		}
		tickTimer := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-tickTimer.C:
				//fmt.Println("发送心跳包")
				pingReqPack := protocolPacket.NewPingReqPacket()
				pingReqPack.Write(conn)
			}
		}
	}()
	time.Sleep(1 * time.Second) //等待一下连接建立
	for {
		c.ReadPacket(conn)

	}
}

func (c *Consumer)ReadPacket(r net.Conn) (protocolPacket.ControlPacket, error) {
	// 读取数据包
	var fh protocolPacket.FixedHeader
	if err := fh.Read(r); err != nil {
		fmt.Errorf("读取包头失败", err)
	}

	var packet protocolPacket.ControlPacket
	switch protocolPacket.DecodePacketType(fh.TypeAndReserved) {
	case byte(protocol.SUBACK):
		var subAckPacket protocolPacket.SubAckPacket
		err := subAckPacket.Read(r, fh)
		if err != nil {
			fmt.Println("客户端订阅主题失败:等待subAck")
			r.Close()
		}
	case byte(protocol.UNSUBACK):
		var fh protocolPacket.FixedHeader
		var unSubAckPacket protocolPacket.UnSubAckPacket
		_ = fh.Read(r)
		unSubAckPacket.Read(r, fh)
		fmt.Println("收到服务端确认取消订阅")
	case byte(protocol.PINGRESP):
		fmt.Println("收到服务端心跳回应")
	}
	return packet, nil
}

func (c *Consumer) UnSubscribe(conn net.Conn, topic []string, identifier uint16) {
	unSubscribePack := protocolPacket.NewUnSubscribePacket(identifier, topic)
	_ = unSubscribePack.Write(conn)
}

func (c *Consumer) DisConnect(conn net.Conn){
	disConnectPack := protocolPacket.NewDisConnectPacketPacket()
	disConnectPack.Write(conn)
	conn.Close()
}

func consumerNetPacket(topic string, initPosition int64) []byte {
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
		Flag:    common.CH,
		Message: common.Message{},
	}
	sendData, _ := json.Marshal(netPacket)
	return sendData
}
