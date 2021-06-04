package client

import (
	"context"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/protocol"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"github.com/zengzhuozhen/gomq/protocol/utils"
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
	client := newClient(opts)
	ctx, cancel := context.WithCancel(context.Background())
	err := client.connect()
	if err != nil {
		panic("连接服务端失败")
	}
	return &Consumer{client: client, ctx: ctx, cancelFunc: cancel}
}

func (c *Consumer) Subscribe(topic []string, QoSRequire int) <-chan *common.MessageUnit {
	c.topic = topic
	// 连接服务端
	subscribePacket := protocolPacket.NewSubscribePacket(c.client.getAvailableIdentity(), topic, byte(QoSRequire))
	err := subscribePacket.Write(c.client.conn)
	if err != nil {
		log.Errorf("客户端订阅主题失败:发送subscribe")
	}
	MsgUnitChan := make(chan *common.MessageUnit, 1000)

	go c.heart(3 * time.Second)
	go c.readPacket(MsgUnitChan)
	return MsgUnitChan
}

func (c *Consumer) heart(duration time.Duration) {
	tickTimer := time.NewTicker(duration)
	for {
		select {
		case <-tickTimer.C:
			log.Infof("发送心跳包")
			pingReqPack := protocolPacket.NewPingReqPacket()
			pingReqPack.Write(c.client.conn)
		case <-c.ctx.Done():
			log.Infof("停止发送心跳")
			return
		}
	}
}

func (c *Consumer) readPacket(msgUnitChan chan<- *common.MessageUnit) {
	for {
		// 读取数据包
		var fh protocolPacket.FixedHeader
		var packet protocolPacket.ControlPacket

		if err := fh.Read(c.client.conn); err != nil {
			log.Errorf("读取包头失败%+v", err)
			return
		}
		switch utils.DecodePacketType(fh.TypeAndReserved) {
		case byte(protocol.SUBACK):
			packet = &protocolPacket.SubAckPacket{}
			err := packet.Read(c.client.conn, fh)
			if err != nil {
				log.Errorf("客户端订阅主题失败")
				c.client.conn.Close()
			} else {
				log.Infof("收到服务端确认订阅消息")
			}
			c.client.IdentityPool[int(packet.(*protocolPacket.SubAckPacket).PacketIdentifier)] = false
		case byte(protocol.UNSUBACK):
			packet = &protocolPacket.UnSubAckPacket{}
			err := packet.Read(c.client.conn, fh)
			if err != nil {
				log.Errorf("客户端取消订阅主题失败")
			} else {
				log.Infof("收到服务端确认取消订阅")
			}
			c.client.IdentityPool[int(packet.(*protocolPacket.UnSubAckPacket).PacketIdentifier)] = false
		case byte(protocol.PINGRESP):
			log.Infof("收到服务端心跳回应")
		default:
			// 普通消息
			messByte := make([]byte, 4096 * 1024) // todo fix:这里可能由于粘包导致超出slice长度
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
	unSubscribePack := protocolPacket.NewUnSubscribePacket(c.client.getAvailableIdentity(), topic)
	_ = unSubscribePack.Write(c.client.conn)
}

func (c *Consumer) Close() {
	c.cancelFunc()
	c.client.disConnect()
}
