package client

import (
	"testing"
)

func TestConsumer_Subscribe_Topic_A(t *testing.T) {
	consumer := NewConsumer("tcp","127.0.0.1",9000,50)
	consumer.Subscribe("A")
}

func TestConsumer_Subscribe_Topic_B(t *testing.T) {
	consumer := NewConsumer("tcp","127.0.0.1",9000,50)
	consumer.Subscribe("B")
}

func TestConsumer_Subscribe_Topic_C(t *testing.T) {
	consumer := NewConsumer("tcp","127.0.0.1",9000,50)
	consumer.Subscribe("C")
}

