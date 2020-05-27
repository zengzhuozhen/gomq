package client

import (
	"gomq/common"
	"testing"
)

func TestNewProducer(t *testing.T) {
	NewProducer(make(chan common.Message,10))
}

func TestProducer_Produce(t *testing.T) {
	producer := NewProducer(make(chan common.Message,10))
	msg := common.NewMessage(1,"hello","hello world")
	producer.Produce(msg)
}
