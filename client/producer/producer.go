package producer

import (
	"fmt"
	"gomq/client"
	"gomq/common"
	"gomq/protocol"
	"gomq/protocol/packet"
	"log"
	"net"
)

type IProducer interface {
	Publish(topic string, mess common.Message, QoS int)
}

type Producer struct {
	bc *client.BasicClient
}

func NewProducer(protocol, host string, port, timeout int) IProducer {
	return &Producer{bc: &client.BasicClient{
		Protocol: protocol,
		Host:     host,
		Port:     port,
		Timeout:  timeout,
	}}
}

func (p *Producer) Publish(topic string, mess common.Message, QoS int) {
	conn, err := p.bc.Connect()
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
	err = publishPacket.Write(conn)
	if err != nil {
		log.Fatal(err.Error())
	}

	switch QoS {
	case protocol.AtMostOnce :
		// noting to do
	case protocol.AtLeastOnce:
		p.WaitAck(conn)
		fmt.Println("读取到puback，完成publish")
	case protocol.ExactOnce:
		p.WaitRec(conn)
	}
	conn.Close()
	return


}

func (p *Producer) WaitAck(r net.Conn) {
	var fh packet.FixedHeader
	if err := fh.Read(r);err != nil{
		fmt.Errorf("读取包头失败,%+v",err)
	}

	pubAckPacket := packet.PubAckPacket{}
	pubAckPacket.Read(r,fh)

	p.bc.IdentityPool[int(pubAckPacket.PacketIdentifier)] = true
	return
}



func (p *Producer) WaitRec(conn net.Conn) {
	//等待 rec
	var fh packet.FixedHeader
	var pubRecPacket packet.PubRecPacket

	if err := fh.Read(conn);err !=nil{
		fmt.Println("接收pubRec包头内容错误",err)
	}
	_ = pubRecPacket.Read(conn, fh)
	fmt.Println("收到rec")
	// PUBREL – 发布释放（QoS 2，第二步)
	pubRelPacket := packet.NewPubRelPacket(pubRecPacket.PacketIdentifier)
	_ = pubRelPacket.Write(conn)

	// 等待 comp
	var pubCompPacket packet.PubCompPacket
	fh = *new(packet.FixedHeader)
	if err := fh.Read(conn);err !=nil{
		fmt.Println("接收pubComp包头内容错误",err)
	}
	_ = pubCompPacket.Read(conn,fh)
	fmt.Println("收到comp")
	p.bc.IdentityPool[int(pubCompPacket.PacketIdentifier)] = true
	return
}


func (p *Producer) GetAvailableIdentity() uint16 {
	for k, v := range p.bc.IdentityPool {
		if v == true {
			p.bc.IdentityPool[k] = false
			return uint16(k)
		}
	}
	return 0
}