package client

import (
	"testing"
)

func TestNewConsumer(t *testing.T) {
	NewConsumer("tcp","127.0.0.1",9000,5)
}

func TestConsumer_Subscribe(t *testing.T) {
	consumer := NewConsumer("tcp","127.0.0.1",9000,5)
	consumer.Subscribe()
}


