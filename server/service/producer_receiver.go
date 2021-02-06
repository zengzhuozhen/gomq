package service

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/protocol"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"github.com/zengzhuozhen/gomq/protocol/visit"
	"github.com/zengzhuozhen/gomq/server/store"
	"net"
)

type ProducerReceiver struct {
	Queue           *common.Queue
	RetainQueue     *common.RetainQueue
	tempPublishPool map[int]protocolPacket.PublishPacket
}

func NewProducerReceiver(queue *common.Queue, store store.Store) *ProducerReceiver {
	return &ProducerReceiver{
		Queue:           queue,
		RetainQueue:     common.NewRetainQueue(store.ReadAll, store.Reset, store.Cap),
		tempPublishPool: make(map[int]protocolPacket.PublishPacket),
	}
}

func (p *ProducerReceiver) ProduceAndResponse(conn net.Conn, publishPacket *protocolPacket.PublishPacket) {
	_ = visit.NewPublishPacketVisitor(&visit.PacketVisitor{Packet: publishPacket}, p.RetainQueue).
		Visit(func(packet protocolPacket.ControlPacket) error {
			switch publishPacket.QoS() {
			case protocol.AtMostOnce:
				p.toQueue(publishPacket) // 最多一次，直接存
			case protocol.AtLeastOnce:
				p.toQueue(publishPacket) // 最少一次,直接存
				p.responsePubAck(conn, publishPacket.PacketIdentifier)
			case protocol.ExactOnce: // 精确一次，这里先不存
				p.responsePubRec(conn, publishPacket)
			}
			return nil
		})
}

func (p *ProducerReceiver) toQueue(publishPacket *protocolPacket.PublishPacket) {
	message := new(common.Message)
	message = message.UnPack(publishPacket.Payload)
	messageUnit := common.NewMessageUnit(publishPacket.TopicName, publishPacket.QoS(), *message)
	log.Debugf("主题 %s 生产了: %s ", publishPacket.TopicName, message.MsgKey)
	p.Queue.Push(messageUnit)
	if publishPacket.Retain() {
		p.RetainQueue.Push(messageUnit)
	}
	log.Debugf("记录入队数据", message.MsgKey)
}

func (p *ProducerReceiver) responsePubAck(conn net.Conn, identify uint16) {
	fmt.Println("发送puback")
	pubAckPacket := protocolPacket.NewPubAckPacket(identify)
	pubAckPacket.Write(conn)
}

func (p *ProducerReceiver) responsePubRec(conn net.Conn, packet *protocolPacket.PublishPacket) {
	log.Debugf("准备返回pubRec")
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
	log.Debugf("准备返回pubComp")
	pubCompPacket := protocolPacket.NewPubCompPacket(packet.PacketIdentifier)
	pubCompPacket.Write(conn)
}
