package client

import (
	"fmt"
	"testing"
)

func TestConsumer_Subscribe_Topic_A(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A"}
	retChan := consumer.Subscribe(topicList)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func TestConsumer_Subscribe_Topic_B(t *testing.T) {
	opts := defaultOpts()

	consumer := NewConsumer(opts)
	topicList := []string{"B"}
	retChan := consumer.Subscribe(topicList)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func TestConsumer_Subscribe_Topic_C(t *testing.T) {
	opts := defaultOpts()

	consumer := NewConsumer(opts)
	topicList := []string{"C"}
	retChan := consumer.Subscribe(topicList)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func TestConsumer_Subscribe_Topic_A_B_C(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A", "B", "C"}
	retChan := consumer.Subscribe(topicList)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func defaultOpts() *Option {
	return &Option{
		Protocol: "tcp",
		Host:     "127.0.0.1",
		Port:     9000,
		Timeout:  50,
	}
}
