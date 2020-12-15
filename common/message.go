package common

import (
	"gomq/protocol/utils"
)

type Message struct {
	MsgId  uint64 `json:"msg_id"`
	MsgKey string `json:"msg_key"`
	Body   string `json:"body"`
}

func (m *Message) Pack() []byte {
	return utils.StructToBytes(m)
}

func (m *Message) UnPack(b []byte) *Message {
	return utils.BytesToStruct(b, &Message{}).(*Message)
}

// 消息单元
type MessageUnit struct {
	Topic string
	QoS   byte
	Data  Message
}

func (m *MessageUnit) Pack() []byte {
	return utils.StructToBytes(m)
}

func (m *MessageUnit) UnPack(b []byte) *MessageUnit {
	return utils.BytesToStruct(b, &MessageUnit{}).(*MessageUnit)
}

type MsgUnitChan chan MessageUnit

func NewMessageUnit(topic string, QoS byte, data Message) MessageUnit {
	return MessageUnit{
		Topic: topic,
		QoS:   QoS,
		Data:  data,
	}
}
