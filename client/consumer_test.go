package client

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/protocol"
	"testing"
	"time"
)


func TestConsumer_Subscribe_Topic_A(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A"}
	retChan := consumer.Subscribe(topicList,protocol.AtMostOnce)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func TestConsumer_Subscribe_Topic_B(t *testing.T) {
	opts := defaultOpts()

	consumer := NewConsumer(opts)
	topicList := []string{"B"}
	retChan := consumer.Subscribe(topicList,protocol.AtMostOnce)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func TestConsumer_Subscribe_Topic_C(t *testing.T) {
	opts := defaultOpts()

	consumer := NewConsumer(opts)
	topicList := []string{"C"}
	retChan := consumer.Subscribe(topicList,protocol.AtMostOnce)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func TestConsumer_Subscribe_Topic_A_B_C(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A", "B", "C"}
	retChan := consumer.Subscribe(topicList,protocol.AtMostOnce)
	for msg := range retChan {
		fmt.Println(msg)
	}
}

func TestConsumer_Subscribe_Then_Unsubscribe(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A"}
	retChan := consumer.Subscribe(topicList,protocol.AtMostOnce)
	go func() {
		for msg := range retChan {
			fmt.Println(msg)
		}
	}()
	time.Sleep(5 * time.Second)
	consumer.UnSubscribe(topicList)
	time.Sleep(15 * time.Second)
}

func TestConsumer_Subscribe_Then_Close(t *testing.T) {
	opts := defaultOpts()
	consumer := NewConsumer(opts)
	topicList := []string{"A", "B", "C"}
	retChan := consumer.Subscribe(topicList,protocol.AtMostOnce)
	go func() {
		for msg := range retChan {
			fmt.Println(msg)
		}
	}()
	time.Sleep(5 * time.Second)
	consumer.Close()
	time.Sleep(5 * time.Second)
}

func defaultOpts() *Option {
	return &Option{
		Protocol: "tcp",
		Address:  "127.0.0.1:9000",
		KeepAlive:  10,
	}
}
