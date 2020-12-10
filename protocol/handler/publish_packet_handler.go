package handler

import (
	"gomq/protocol/packet"
)

type PublishPacketHandler interface {
	HandleAll()
	// 重发标识
	HandleDup()
	// 保留标识
	HandleRetain()
	// 服务质量等级
	ReturnQoS() int
}

type PublishPacketHandle struct {
	publishPacket *packet.PublishPacket
	returnQoS int
}



func (handler *PublishPacketHandle) HandleAll() {
	Byte4 := handler.publishPacket.TypeAndReserved - (16 + 32)
	if Byte4 >= 8 {
		handler.HandleDup()
		Byte4 -= 8
	}

	if Byte4 % 2 == 1 {
		handler.HandleRetain()
		Byte4--
	}
	handler.returnQoS = int(Byte4 >> 1) // 服务质量等级 QoS,右移一位转为 Enum值
}

func (handler *PublishPacketHandle) HandleDup() {
	panic("implement me")
}


// 如果客户端发给服务端的PUBLISH报文的保留（RETAIN）标志被设置为1，服务端必须存储这个应用消息和它的服务质量等级（QoS）
//，以便它可以被分发给未来的主题名匹配的订阅者 [MQTT-3.3.1-5]。
//一个新的订阅建立时，对每个匹配的主题名，如果存在最近保留的消息，它必须被发送给这个订阅者 [MQTT-3.3.1-6]。
//如果服务端收到一条保留（RETAIN）标志为1的QoS 0消息，它必须丢弃之前为那个主题保留的任何消息。
//它应该将这个新的QoS 0消息当作那个主题的新保留消息，但是任何时候都可以选择丢弃它 — 如果这种情况发生了，那个主题将没有保留消息 [MQTT-3.3.1-7]。
func (handler *PublishPacketHandle) HandleRetain() {
	panic("implement me")
	if handler.publishPacket.Retain() {
		// todo 保存应用消息和服务质量等级
	}

	if handler.publishPacket.Retain() == false && handler.publishPacket.QoS() == 0 {
		// todo 丢弃之前保留的主题信息
	}
}

func (handler *PublishPacketHandle) ReturnQoS() int {
	return handler.returnQoS
}

func NewPublishPacketHandle(publishPacket *packet.PublishPacket) PublishPacketHandler {
	return  &PublishPacketHandle{publishPacket: publishPacket}
}
