package common

import (
	"github.com/google/uuid"
	"github.com/zengzhuozhen/gomq/protocol/utils"
)

type Message struct {
	Id     string
	MsgKey string
	Body   string
}

func (m *Message) Pack() []byte {
	m.Id = uuid.New().String()
	return utils.StructToBytes(m)
}

func (m *Message) UnPack(b []byte) *Message {
	return utils.BytesToStruct(b, &Message{}).(*Message)
}

func (m *Message) GetId() string {
	return m.Id
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
