package client

import (
	"context"
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/protocol"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"log"
	"time"
)

type Producer struct {
	client                  *client
	cancelResendPublishFunc context.CancelFunc
	cancelResendPubrelFunc  context.CancelFunc
}

func NewProducer(opts *Option) *Producer {
	client := newClient(opts)
	err := client.connect()
	if err != nil {
		panic("连接服务端失败")
	}
	return &Producer{client: client}
}

func (p *Producer) Publish(messageUnit common.MessageUnit, QoS, retain int) {
	var identity uint16
	if QoS != protocol.AtMostOnce {
		identity = p.client.getAvailableIdentity()
	}

	publishPacket := protocolPacket.NewPublishPacket(messageUnit.Topic, messageUnit.Data, true, QoS, retain, identity)
	err := publishPacket.Write(p.client.conn)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	p.cancelResendPublishFunc = cancelFunc

	switch QoS {
	case protocol.AtMostOnce:
		// noting to do
	case protocol.AtLeastOnce:
		resendPacket := protocolPacket.NewPublishPacket(messageUnit.Topic, messageUnit.Data, false, QoS, retain, identity)
		go p.overtimeResendPublish(ctx, resendPacket)
		p.client.session.save(&resendPacket) // 如果断线重发，应该采用重发的包
		p.WaitAck()
	case protocol.ExactOnce:
		resendPacket := protocolPacket.NewPublishPacket(messageUnit.Topic, messageUnit.Data, false, QoS, retain, identity)
		go p.overtimeResendPublish(ctx, resendPacket)
		p.client.session.save(&resendPacket) // 同上
		p.WaitRecAndComp()
	}
	return
}

func (p *Producer) WaitAck() {
	var fh protocolPacket.FixedHeader
	if err := fh.Read(p.client.conn); err != nil {
		fmt.Errorf("读取包头失败,%+v", err)
	}
	pubAckPacket := protocolPacket.PubAckPacket{}
	pubAckPacket.Read(p.client.conn, fh)

	p.client.IdentityPool[int(pubAckPacket.PacketIdentifier)] = true
	fmt.Println("读取到puback，完成publish")
	p.client.session.remove(&pubAckPacket)
	p.cancelResendPublishFunc()
	return
}

func (p *Producer) WaitRecAndComp() {
	//等待 rec
	var fh protocolPacket.FixedHeader

	for {
		if err := fh.Read(p.client.conn); err != nil {
			fmt.Println("接收数据包头内容错误", err)
		}
		var packet protocolPacket.ControlPacket

		switch utils.DecodePacketType(fh.TypeAndReserved) {
		case byte(protocol.PUBREC):
			packet = &protocolPacket.PubRecPacket{}
			_ = packet.Read(p.client.conn, fh)
			p.client.IdentityPool[int(packet.(*protocolPacket.PubRecPacket).PacketIdentifier)] = true
			p.client.session.save(packet)
			p.client.session.remove(packet)

			// PUBREL – 发布释放（QoS 2，第二步)
			pubRelPacket := protocolPacket.NewPubRelPacket(packet.(*protocolPacket.PubRecPacket).PacketIdentifier)
			_ = pubRelPacket.Write(p.client.conn)
			p.cancelResendPublishFunc() // 不再发布PUBLISH
			// 重发PUBREL逻辑
			ctx, cancelFunc := context.WithCancel(context.Background())
			p.cancelResendPubrelFunc = cancelFunc
			p.client.IdentityPool[int(packet.(*protocolPacket.PubRecPacket).PacketIdentifier)] = false
			p.client.session.save(&pubRelPacket)
			p.client.session.remove(&pubRelPacket)
			go p.overtimeResendPubrel(ctx, pubRelPacket)

		case byte(protocol.PUBCOMP):
			packet = &protocolPacket.PubCompPacket{}
			_ = packet.Read(p.client.conn, fh)
			fmt.Println("收到comp,完成publish")
			p.cancelResendPubrelFunc()
			p.client.IdentityPool[int(packet.(*protocolPacket.PubCompPacket).PacketIdentifier)] = true
			p.client.session.remove(packet)
			return
		}
	}
}

// 超时重发Publish包
func (p *Producer) overtimeResendPublish(ctx context.Context, publishPacket protocolPacket.PublishPacket) {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			if p.client.IdentityPool[int(publishPacket.PacketIdentifier)] == true {
				return
			}
			err := publishPacket.Write(p.client.conn)
			if err != nil {
				log.Fatal(err.Error())
			}
		case <-ctx.Done():
			// 用于 QoS2 第二阶段，发送了 PUBREL 报文就不能重发这个PUBLISH报文
			return
		}
	}
}

// 超时重发Pubrel包
func (p *Producer) overtimeResendPubrel(ctx context.Context, pubrelPacket protocolPacket.PubRelPacket) {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			if p.client.IdentityPool[int(pubrelPacket.PacketIdentifier)] == true {
				return
			}
			err := pubrelPacket.Write(p.client.conn)
			if err != nil {
				log.Fatal(err.Error())
			}
		case <-ctx.Done():
			// 用于 QoS2 第三阶段，接收了comp包就不再重发PUBREL包
			return
		}
	}
}

func (p *Producer) Close() {
	p.client.disConnect()
}
