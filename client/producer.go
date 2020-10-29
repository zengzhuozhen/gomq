package client

import (
	"fmt"
	"gomq/common"
	"gomq/protocol"
	"gomq/protocol/packet"
	"log"
	"sync"
)

type IProducer interface {
	Publish(topic string, mess common.Message, QoS int)
	WaitAck()
	WaitRec()
	GetAvailableIdentity() uint16
}

type Producer struct {
	client *client
}

func NewProducer(opts *Option) IProducer {
	return &Producer{client: &client{
		options:      opts,
		optionsMu:    sync.Mutex{},
	}}
}

func (p *Producer) Publish(topic string, mess common.Message, QoS int) {
	err := p.client.Connect()
	if err != nil {
		panic("连接服务端失败")
	}

	// todo retain暂时设置为0,后期优化
	var identity uint16
	if QoS != protocol.AtMostOnce {
		identity = p.GetAvailableIdentity()
	}
	publishPacket := packet.NewPublishPacket(topic, mess, true, QoS, 0, identity)
	//_, err = conn.Write(utils.StructToBytes(publishPacket))
	err = publishPacket.Write(p.client.conn)
	if err != nil {
		log.Fatal(err.Error())
	}

	switch QoS {
	case protocol.AtMostOnce:
		// noting to do
	case protocol.AtLeastOnce:
		p.WaitAck()
		fmt.Println("读取到puback，完成publish")
	case protocol.ExactOnce:
		p.WaitRec()
	}
	p.client.conn.Close()
	return

}

func (p *Producer) WaitAck() {
	var fh packet.FixedHeader
	if err := fh.Read(p.client.conn); err != nil {
		fmt.Errorf("读取包头失败,%+v", err)
	}

	pubAckPacket := packet.PubAckPacket{}
	pubAckPacket.Read(p.client.conn, fh)

	p.client.IdentityPool[int(pubAckPacket.PacketIdentifier)] = true
	return
}

func (p *Producer) WaitRec() {
	//等待 rec
	var fh packet.FixedHeader
	var pubRecPacket packet.PubRecPacket

	if err := fh.Read(p.client.conn); err != nil {
		fmt.Println("接收pubRec包头内容错误", err)
	}
	_ = pubRecPacket.Read(p.client.conn, fh)
	fmt.Println("收到rec")
	// PUBREL – 发布释放（QoS 2，第二步)
	pubRelPacket := packet.NewPubRelPacket(pubRecPacket.PacketIdentifier)
	_ = pubRelPacket.Write(p.client.conn)

	// 等待 comp
	var pubCompPacket packet.PubCompPacket
	fh = *new(packet.FixedHeader)
	if err := fh.Read(p.client.conn); err != nil {
		fmt.Println("接收pubComp包头内容错误", err)
	}
	_ = pubCompPacket.Read(p.client.conn, fh)
	fmt.Println("收到comp")
	p.client.IdentityPool[int(pubCompPacket.PacketIdentifier)] = true
	return
}

func (p *Producer) GetAvailableIdentity() uint16 {
	for k, v := range p.client.IdentityPool {
		if v == true {
			p.client.IdentityPool[k] = false
			return uint16(k)
		}
	}
	return 0
}
