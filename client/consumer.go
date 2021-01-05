package client

import (
	"context"
	"fmt"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"strings"
	"time"
)



type Consumer struct {
	client     *client
	ctx        context.Context
	cancelFunc context.CancelFunc
	topic      []string
}

func NewConsumer(opts *Option) *Consumer {
	client := NewClient(opts)
	ctx, cancel := context.WithCancel(context.Background())
	err := client.Connect()
	if err != nil {
		panic("连接服务端失败")
	}
	return &Consumer{client: client, ctx: ctx, cancelFunc: cancel}
}

func (c *Consumer) Subscribe(topic []string) <-chan *common.MessageUnit {
	c.topic = topic
	// 连接服务端
	subscribePacket := protocolPacket.NewSubscribePacket(0, topic, 0)
	err := subscribePacket.Write(c.client.conn)
	if err != nil {
		fmt.Println("客户端订阅主题失败:发送subscribe")
	}
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
		case <-c.ctx.Done():
			fmt.Println("停止发送心跳")
			return
		}
	}
}

func (c *Consumer) ReadPacket(msgUnitChan chan<- *common.MessageUnit) {
	for {
		// 读取数据包
		var fh protocolPacket.FixedHeader
		var packet protocolPacket.ControlPacket

		if err := fh.Read(c.client.conn); err != nil {
			fmt.Printf("读取包头失败%+v", err)
			return
		}
		switch utils.DecodePacketType(fh.TypeAndReserved) {
		case byte(protocol.SUBACK):
			packet = &protocolPacket.SubAckPacket{}
			err := packet.Read(c.client.conn, fh)
			if err != nil {
				fmt.Println("客户端订阅主题失败")
				c.client.conn.Close()
			} else {
				fmt.Println("收到服务端确认订阅消息")
			}
		case byte(protocol.UNSUBACK):
			packet = &protocolPacket.UnSubAckPacket{}
			err := packet.Read(c.client.conn, fh)
			if err != nil {
				fmt.Println("客户端取消订阅主题失败")
			} else {
				fmt.Println("收到服务端确认取消订阅")
			}
		case byte(protocol.PINGRESP):
			fmt.Println("收到服务端心跳回应")
		default:
			// 普通消息
			messByte := make([]byte, 4096) // todo fix:这里可能由于粘包导致超出slice长度
			n, _ := c.client.conn.Read(messByte)
			head := fh.Pack()
			data := append(head.Bytes(), messByte[:n]...)
			byteList := strings.Split(string(data), "\n")
			for _, byteItem := range byteList[:len(byteList)-1] {
				msg := new(common.MessageUnit)
				msgUnitChan <- msg.UnPack([]byte(byteItem))
			}
		}
	}
}

func (c *Consumer) UnSubscribe(topic []string) {
	unSubscribePack := protocolPacket.NewUnSubscribePacket(c.client.GetAvailableIdentity(), topic)
	_ = unSubscribePack.Write(c.client.conn)
}

func (c *Consumer) Close() {
	c.cancelFunc()
	c.client.DisConnect()
}
