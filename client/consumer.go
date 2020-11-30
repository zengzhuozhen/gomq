package client

import (
	"fmt"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"time"
)

type IConsumer interface {
	Subscribe(topic []string) <-chan *common.MessageUnit
	Heart(duration time.Duration)
	ReadPacket(msgChan chan<- *common.MessageUnit)
	UnSubscribe(topic []string, identifier uint16)
	DisConnect()
}

type Consumer struct {
	client *client
}

func NewConsumer(opts *Option) IConsumer {
	client := NewClient(opts).(*client)
	return &Consumer{client: client}
}

func (c *Consumer) Subscribe(topic []string) <-chan *common.MessageUnit {
	err := c.client.Connect()
	if err != nil {
		panic("连接服务端失败")
	}
	// 连接服务端
	subscribePacket := protocolPacket.NewSubscribePacket(1, topic, 0)
	err = subscribePacket.Write(c.client.conn)
	if err != nil {
		fmt.Println("客户端订阅主题失败:发送subscribe")
	}
	fmt.Println("发送subscribe")

	MsgUnitChan := make(chan *common.MessageUnit, 1000)
	go c.Heart(3 * time.Second)
	go c.ReadPacket(MsgUnitChan)
	return MsgUnitChan
}

func (c *Consumer) Heart(duration time.Duration) {
	tickTimer := time.NewTicker(duration)
	for {
		select {
		case <-tickTimer.C:
			fmt.Println("发送心跳包")
			pingReqPack := protocolPacket.NewPingReqPacket()
			pingReqPack.Write(c.client.conn)
		}
	}
}

func (c *Consumer) ReadPacket(msgUnitChan chan<- *common.MessageUnit) {
	for {
		// 读取数据包
		var fh protocolPacket.FixedHeader
		if err := fh.Read(c.client.conn); err != nil {
			fmt.Errorf("读取包头失败%+v", err)
			return
		}

		switch utils.DecodePacketType(fh.TypeAndReserved) {
		case byte(protocol.SUBACK):
			var subAckPacket protocolPacket.SubAckPacket
			err := subAckPacket.Read(c.client.conn, fh)
			if err != nil {
				fmt.Println("客户端订阅主题失败:等待subAck")
				c.client.conn.Close()
			}
			fmt.Println("收到服务端确认订阅消息")
		case byte(protocol.UNSUBACK):
			var fh protocolPacket.FixedHeader
			var unSubAckPacket protocolPacket.UnSubAckPacket
			_ = fh.Read(c.client.conn)
			unSubAckPacket.Read(c.client.conn, fh)
			fmt.Println("收到服务端确认取消订阅")
		case byte(protocol.PINGRESP):
			fmt.Println("收到服务端心跳回应")
		default:
			// 普通消息
			messByte := make([]byte, 4096) // todo fix:这里可能由于粘包导致超出slice长度
			n, _ := c.client.conn.Read(messByte)
			head := fh.Pack()
			data := append(head.Bytes(), messByte[:n]...)
			message := new(common.MessageUnit)
			message = message.UnPack(data)
			msgUnitChan <- message
		}
	}
}

func (c *Consumer) UnSubscribe(topic []string, identifier uint16) {
	unSubscribePack := protocolPacket.NewUnSubscribePacket(identifier, topic)
	_ = unSubscribePack.Write(c.client.conn)
}

func (c *Consumer) DisConnect() {
	disConnectPack := protocolPacket.NewDisConnectPacketPacket()
	_ = disConnectPack.Write(c.client.conn)
	_ = c.client.conn.Close()
}
