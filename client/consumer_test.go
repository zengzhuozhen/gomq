package client

import (
	"testing"
)

func TestConsumer_Subscribe_Topic_A(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A"}
	consumer.Subscribe(topicList)
}

func TestConsumer_Subscribe_Topic_B(t *testing.T) {
	opts := defaultOpts()

	consumer := NewConsumer(opts)
	topicList := []string{"B"}
	consumer.Subscribe(topicList)
}

func TestConsumer_Subscribe_Topic_C(t *testing.T) {
	opts := defaultOpts()

	consumer := NewConsumer(opts)
	topicList := []string{"C"}
	consumer.Subscribe(topicList)
}

func TestConsumer_Subscribe_Topic_A_B_C(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A", "B", "C"}
	consumer.Subscribe(topicList)
}

func defaultOpts() *Option {
	return &Option{
		Protocol: "tcp",
		Host:     "127.0.0.1",
		Port:     9000,
		Timeout:  50,
	}
}
