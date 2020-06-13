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
	return  utils.BytesToStruct(b, &Message{}).(*Message)
}

type MsgChan chan Message
