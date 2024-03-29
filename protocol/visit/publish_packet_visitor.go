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



func NewPublishPacketVisitor(visitor packet.Visitor,queue *common.RetainQueue) *PublishPacketVisitor {
	container := container{retainQueue: queue}
	return &PublishPacketVisitor{
		filterVisitor: packet.NewFilteredVisitor(visitor,
			container.qosValidate,
			container.handleDup,
			container.handleRetain,
		),
	}
}

func (c container)qosValidate(controlPacket packet.ControlPacket) error {
	publishPacket := controlPacket.(*packet.PublishPacket)
	if publishPacket.QoS() < 0 || publishPacket.QoS() > 2 {
		return fmt.Errorf("非法的Qos")
	}
	return nil
}

func (c container)handleDup(controlPacket packet.ControlPacket) error {
	publishPacket := controlPacket.(*packet.PublishPacket)
	if publishPacket.Dup() {
		// 重发报文，暂时没有处理
	}
	return nil
}

func (c container) handleRetain(controlPacket packet.ControlPacket) error {
	publishPacket := controlPacket.(*packet.PublishPacket)
	if publishPacket.Retain() == true  {
		if publishPacket.QoS() == 0{
			// 丢弃之前保留的主题信息,将该消息作为最新的保留消息
			c.retainQueue.Reset(publishPacket.TopicName)
		}
		message := new(common.Message)
		message = message.UnPack(publishPacket.Payload)
		messageUnit := common.NewMessageUnit(publishPacket.TopicName, publishPacket.QoS(), *message)
		c.retainQueue.Push(messageUnit)
	}
	return nil
}
