package visit

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/protocol/packet"
)

type PublishPacketVisitor struct {
	filterVisitor packet.Visitor
}

func (v *PublishPacketVisitor) Visit(fn packet.VisitorFunc) error {
	return v.filterVisitor.Visit(fn)
}

type container struct {
	retainQueue *common.RetainQueue
}

func NewPublishPacketVisitor(visitor packet.Visitor,queue *common.RetainQueue) *PublishPacketVisitor {
	rqc := container{retainQueue: queue}
	return &PublishPacketVisitor{
		filterVisitor: packet.NewFilteredVisitor(visitor,
			qosValidate,
			handleDup,
			rqc.handleRetain,
		),
	}
}

func qosValidate(controlPacket packet.ControlPacket) error {
	publishPacket := controlPacket.(*packet.PublishPacket)
	if publishPacket.QoS() >= 3 {
		return fmt.Errorf("非法的Qos")
	}
	return nil
}

func handleDup(controlPacket packet.ControlPacket) error {
	publishPacket := controlPacket.(*packet.PublishPacket)
	if publishPacket.Dup() {
		// todo

	}
	return nil
}

func (rqc container) handleRetain(controlPacket packet.ControlPacket) error {
	publishPacket := controlPacket.(*packet.PublishPacket)
	if publishPacket.Retain() == true && publishPacket.QoS() == 0 {
		// 丢弃之前保留的主题信息
		rqc.retainQueue.Reset(publishPacket.TopicName)
		// 将该消息作为最新的保留消息
		message := new(common.Message)
		message = message.UnPack(publishPacket.Payload)
		messageUnit := common.NewMessageUnit(publishPacket.TopicName, publishPacket.QoS(), *message)
		rqc.retainQueue.Push(messageUnit)
	}
	return nil
}
