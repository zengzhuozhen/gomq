package handler

import (
	"gomq/common"
	"gomq/protocol/packet"
)

type PublishPacketHandler interface {
	CheckAll()
	// 重发标识
	HandleDup()
	// 保留标识
	HandleRetain(retainQueue *common.RetainQueue)

	ReturnDup() bool
	ReturnQoS() int
	ReturnRetain() bool
}

type PublishPacketHandle struct {
	publishPacket *packet.PublishPacket
	returnDup     bool
	returnQoS     int
	returnRetain  bool
}

func (handler *PublishPacketHandle) CheckAll() {
	Byte4 := handler.publishPacket.TypeAndReserved - (16 + 32)
	if Byte4 >= 8 {
		handler.returnDup = true
		Byte4 -= 8
	}

	if Byte4%2 == 1 {
		handler.returnRetain = true
		Byte4--
	}
	handler.returnQoS = int(Byte4 >> 1) // 服务质量等级 QoS,右移一位转为 Enum值
}

func (handler *PublishPacketHandle) HandleDup() {
	// todo
}

// 如果客户端发给服务端的PUBLISH报文的保留（RETAIN）标志被设置为1，服务端必须存储这个应用消息和它的服务质量等级（QoS）
//，以便它可以被分发给未来的主题名匹配的订阅者 [MQTT-3.3.1-5]。
//一个新的订阅建立时，对每个匹配的主题名，如果存在最近保留的消息，它必须被发送给这个订阅者 [MQTT-3.3.1-6]。
//如果服务端收到一条保留（RETAIN）标志为1的QoS 0消息，它必须丢弃之前为那个主题保留的任何消息。
//它应该将这个新的QoS 0消息当作那个主题的新保留消息，但是任何时候都可以选择丢弃它 — 如果这种情况发生了，那个主题将没有保留消息 [MQTT-3.3.1-7]。
func (handler *PublishPacketHandle) HandleRetain(retainQueue *common.RetainQueue) {
	if handler.publishPacket.Retain() == true && handler.publishPacket.QoS() == 0 {
		// 丢弃之前保留的主题信息
		delete(retainQueue.Local, handler.publishPacket.TopicName)
		// 将该消息作为最新的保留消息
		message := new(common.Message)
		message = message.UnPack(handler.publishPacket.Payload)
		messageUnit := common.NewMessageUnit(handler.publishPacket.TopicName, handler.publishPacket.QoS(), *message)
		retainQueue.Local[handler.publishPacket.TopicName] = []common.MessageUnit{messageUnit}
	}
}

func (handler *PublishPacketHandle) ReturnQoS() int {
	return handler.returnQoS
}

func (handler *PublishPacketHandle) ReturnDup() bool {
	return handler.returnDup
}

func (handler *PublishPacketHandle) ReturnRetain() bool {
	return handler.returnRetain
}

func NewPublishPacketHandle(publishPacket *packet.PublishPacket) PublishPacketHandler {
	return &PublishPacketHandle{publishPacket: publishPacket}
}
