package client

import (
	"gomq/common"
	"testing"
)

func TestNewConsumer(t *testing.T) {
	 NewConsumer(make(chan common.Message,10))
}
func TestConsumer_Consume(t *testing.T) {
	consumer := NewConsumer(make(chan common.Message,10))
	consumer.Consume()
}