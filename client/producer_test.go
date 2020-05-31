package client

import (
	"gomq/common"
	"testing"
)


func TestProducer_Publish_Topic_A(t *testing.T) {
	producer := 	NewProducer("tcp","127.0.0.1",9000,10)
	mess := common.Message{
		MsgId: 1,
		MsgKey: "test",
		Body:   "hello world A ",
	}
	producer.Publish("A",mess)
}

func TestProducer_Publish_Topic_B(t *testing.T) {
	producer := 	NewProducer("tcp","127.0.0.1",9000,10)
	mess := common.Message{
		MsgId: 2,
		MsgKey: "test",
		Body:   "hello world B",
	}
	producer.Publish("B",mess)
}

func TestProducer_Publish_Topic_C(t *testing.T) {
	producer := 	NewProducer("tcp","127.0.0.1",9000,10)
	mess := common.Message{
		MsgId: 3,
		MsgKey: "test",
		Body:   "hello world C",
	}
	producer.Publish("C",mess)
}