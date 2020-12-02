package client

import (
	"fmt"
	"testing"
	"time"
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

func TestConsumer_Subscribe_Then_Unsubscribe(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A"}
	retChan := consumer.Subscribe(topicList)
	go func() {
		for msg := range retChan {
			fmt.Println(msg)
		}
	}()
	time.Sleep(5 * time.Second)
	consumer.UnSubscribe(topicList)
	time.Sleep(15 * time.Second)
}

func TestConsumer_Subscribe_Then_Disconnect(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A", "B", "C"}
	retChan := consumer.Subscribe(topicList)
	go func() {
		for msg := range retChan {
			fmt.Println(msg)
		}
	}()
	time.Sleep(5 * time.Second)
	consumer.DisConnect()
	time.Sleep(5 * time.Second)
}

func defaultOpts() *Option {
	return &Option{
		Protocol: "tcp",
		Address:  "127.0.0.1:9000",
		Timeout:  50,
	}
}
