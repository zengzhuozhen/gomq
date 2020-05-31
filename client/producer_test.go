package client

import (
	"gomq/common"
	"testing"
)

func TestNewProducer(t *testing.T) {
	NewProducer("tcp","127.0.0.1",9000,10)
}

func TestProducer_Publish(t *testing.T) {
	producer := 	NewProducer("tcp","127.0.0.1",9000,10)
	mess := common.Message{
		MsgId: 1,
		MsgKey: "test",
		Body:   "hello world",
	}
	producer.Publish(mess)
}