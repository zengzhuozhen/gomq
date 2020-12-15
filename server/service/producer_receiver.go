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
	Queue           *common.Queue
	RetainQueue     *common.RetainQueue
	tempPublishPool map[int]protocolPacket.PublishPacket
}

func NewProducerReceiver(queue *common.Queue) *ProducerReceiver {
	return &ProducerReceiver{
		Queue:           queue,
		RetainQueue:     common.NewRetainQueue(),
		tempPublishPool: make(map[int]protocolPacket.PublishPacket),
	}
}

func (p *ProducerReceiver) ProduceAndResponse(conn net.Conn, publishPacket *protocolPacket.PublishPacket) {
	publishPacketHandler := handler.NewPublishPacketHandle(publishPacket)
	publishPacketHandler.CheckAll()
	publishPacketHandler.HandleDup()
	publishPacketHandler.HandleRetain(p.RetainQueue)

	switch publishPacketHandler.ReturnQoS() {
	case protocol.AtMostOnce:
		p.toQueue(publishPacket) // 最多一次，直接存
	case protocol.AtLeastOnce:
		p.toQueue(publishPacket) // 最少一次,直接存
		p.responsePubAck(conn, publishPacket.PacketIdentifier)
	case protocol.ExactOnce: // 精确一次，这里先不存
		p.responsePubRec(conn, publishPacket)
	}
}


func (p *ProducerReceiver) toQueue(publishPacket *protocolPacket.PublishPacket) {
	message := new(common.Message)
	message = message.UnPack(publishPacket.Payload)
	messageUnit := common.NewMessageUnit(publishPacket.TopicName, publishPacket.QoS(), *message)
	fmt.Printf("主题 %s 生产了: %s ", publishPacket.TopicName, message.MsgKey)
	p.Queue.Push(messageUnit)
	fmt.Println("当前列表",p.Queue.Local)
	if publishPacket.Retain() {
		p.RetainQueue.Push(messageUnit)
	}
	fmt.Println("记录入队数据", message.MsgKey)
}

func (p *ProducerReceiver) responsePubAck(conn net.Conn, identify uint16) {
	fmt.Println("发送puback")
	pubAckPacket := protocolPacket.NewPubAckPacket(identify)
	pubAckPacket.Write(conn)
}

func (p *ProducerReceiver) responsePubRec(conn net.Conn, packet *protocolPacket.PublishPacket) {
	fmt.Println("准备返回pubRec")
	// PUBREC – 发布收到（QoS 2，第一步)
	pubRecPacket := protocolPacket.NewPubRecPacket(packet.PacketIdentifier)
	pubRecPacket.Write(conn)
	p.tempPublishPool[int(packet.PacketIdentifier)] = *packet
}

func (p *ProducerReceiver) AcceptRelAndRespComp(conn net.Conn, packet *protocolPacket.PubRelPacket) {
	publishPacket := p.tempPublishPool[int(packet.PacketIdentifier)]
	p.toQueue(&publishPacket)
	delete(p.tempPublishPool, int(packet.PacketIdentifier)) // 删除临时保存的publish包，防止重发
	// PUBCOMP – 发布完成（QoS 2，第三步)
	fmt.Println("准备返回pubComp")
	pubCompPacket := protocolPacket.NewPubCompPacket(packet.PacketIdentifier)
	pubCompPacket.Write(conn)
}
