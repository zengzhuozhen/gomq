package client

import (
	"context"
	"fmt"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"log"
	"time"
)

type IProducer interface {
	Publish(mess common.MessageUnit, QoS int)
	WaitAck()
	WaitRecAndComp(cancelResendFunc context.CancelFunc)
}

type Producer struct {
	client *client
}

func NewProducer(opts *Option) IProducer {
	client := NewClient(opts).(*client)
	return &Producer{client: client}
}

func (p *Producer) Publish(messageUnit common.MessageUnit, QoS int) {
	err := p.client.Connect()
	if err != nil {
		panic("连接服务端失败")
	}

	// todo retain暂时设置为0,后期优化
	var identity uint16
	if QoS != protocol.AtMostOnce {
		identity = p.client.GetAvailableIdentity()
	}
	publishPacket := protocolPacket.NewPublishPacket(messageUnit.Topic, messageUnit.Data, true, QoS, 0, identity)
	err = publishPacket.Write(p.client.conn)
	if err != nil {
		log.Fatal(err.Error())
	}

	switch QoS {
	case protocol.AtMostOnce:
		// noting to do
	case protocol.AtLeastOnce:
		resendPacket := protocolPacket.NewPublishPacket(messageUnit.Topic, messageUnit.Data, false, QoS, 0, identity)
		go p.overtimeAndResend(context.TODO(), resendPacket)
		p.WaitAck()
	case protocol.ExactOnce:
		resendPacket := protocolPacket.NewPublishPacket(messageUnit.Topic, messageUnit.Data, false, QoS, 0, identity)
		ctx, cancelResendFunc := context.WithCancel(context.Background())
		go p.overtimeAndResend(ctx, resendPacket)
		p.WaitRecAndComp(cancelResendFunc)
	}
	p.client.conn.Close()
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
	return
}

func (p *Producer) WaitRecAndComp(cancelResendFunc context.CancelFunc) {
	//等待 rec
	var fh protocolPacket.FixedHeader

	if err := fh.Read(p.client.conn); err != nil {
		fmt.Println("接收数据包头内容错误", err)
	}
	var packet protocolPacket.ControlPacket

	switch utils.DecodePacketType(fh.TypeAndReserved) {
	case byte(protocol.PUBREC):
		packet = &protocolPacket.PubRecPacket{}
		_ = packet.ReadHeadOnly(p.client.conn, fh)
		p.client.IdentityPool[int(packet.(*protocolPacket.PubRecPacket).PacketIdentifier)] = true
		// PUBREL – 发布释放（QoS 2，第二步)
		pubRelPacket := protocolPacket.NewPubRelPacket(packet.(*protocolPacket.PubRecPacket).PacketIdentifier)
		_ = pubRelPacket.Write(p.client.conn)
		cancelResendFunc() // 不再发布PUBLISH
		p.client.IdentityPool[int(packet.(*protocolPacket.PubRecPacket).PacketIdentifier)] = false

	case byte(protocol.PUBCOMP):
		packet = &protocolPacket.PubCompPacket{}
		_ = packet.Read(p.client.conn, fh)
		_ = packet.ReadHeadOnly(p.client.conn, fh)
		var pubCompPacket protocolPacket.PubCompPacket
		fh = *new(protocolPacket.FixedHeader)
		if err := fh.Read(p.client.conn); err != nil {
			fmt.Println("接收pubComp包头内容错误", err)
		}
		_ = pubCompPacket.Read(p.client.conn, fh)
		fmt.Println("收到comp,完成publish")
		p.client.IdentityPool[int(packet.(*protocolPacket.PubCompPacket).PacketIdentifier)] = true
	}
}

// 超时重发
func (p *Producer) overtimeAndResend(ctx context.Context, publishPacket protocolPacket.PublishPacket) {
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
