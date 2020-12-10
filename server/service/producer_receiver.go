package service

import (
	"fmt"
	"gomq/common"
	"gomq/protocol"
	"gomq/protocol/handler"
	protocolPacket "gomq/protocol/packet"
	"net"
)

type ProducerReceiver struct {
	Queue *common.Queue
}

func NewProducerReceiver(queue *common.Queue) *ProducerReceiver {
	return &ProducerReceiver{Queue: queue}
}

func (p *ProducerReceiver) ProduceAndResponse(conn net.Conn, publishPacket *protocolPacket.PublishPacket) {
	publishPacketHandler := handler.NewPublishPacketHandle(publishPacket)
	publishPacketHandler.HandleAll()
	switch publishPacketHandler.ReturnQoS() {
	case protocol.AtMostOnce:
		// nothing to do
	case protocol.AtLeastOnce:
		defer responsePubAck(conn, publishPacket.PacketIdentifier)
	case protocol.ExactOnce:
		defer responsePubRec(conn, publishPacket.PacketIdentifier)
	}
	p.toQueue(publishPacket)
}

func (p *ProducerReceiver) toQueue(publishPacket *protocolPacket.PublishPacket) {
	message := new(common.Message)
	message = message.UnPack(publishPacket.Payload)
	messageUnit := common.NewMessageUnit(publishPacket.TopicName, *message)
	fmt.Printf("主题 %s 生产了: %s ", publishPacket.TopicName, message.MsgKey)
	p.Queue.Push(messageUnit)
	fmt.Println("记录入队数据", message.MsgKey)
}

func responsePubAck(conn net.Conn, identify uint16) {
	fmt.Println("发送puback")
	pubAckPacket := protocolPacket.NewPubAckPacket(identify)
	pubAckPacket.Write(conn)
}

func responsePubRec(conn net.Conn, identify uint16) {
	fmt.Println("准备返回pubRec")
	// PUBREC – 发布收到（QoS 2，第一步)
	pubRecPacket := protocolPacket.NewPubRecPacket(identify)
	pubRecPacket.Write(conn)

	// 等待 rel
	var fh protocolPacket.FixedHeader
	var pubRelPacket protocolPacket.PubRelPacket
	if err := fh.Read(conn); err != nil {
		fmt.Println("接收pubRel包头内容错误", err)
	}
	pubRelPacket.Read(conn, fh)

	// PUBCOMP – 发布完成（QoS 2，第三步)
	fmt.Println("准备返回pubComp")
	pubCompPacket := protocolPacket.NewPubCompPacket(pubRelPacket.PacketIdentifier)
	pubCompPacket.Write(conn)
	// todo identify Pool
	return
}
