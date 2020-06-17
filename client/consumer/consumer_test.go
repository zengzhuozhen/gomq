package consumer

import (
	"testing"
)

func TestConsumer_Subscribe_Topic_A(t *testing.T) {
	consumer := NewConsumer("tcp","127.0.0.1",9000,50)
	topicList := []string{"A"}
	consumer.Subscribe(topicList)
}

func TestConsumer_Subscribe_Topic_B(t *testing.T) {
	consumer := NewConsumer("tcp","127.0.0.1",9000,50)
	topicList := []string{"B"}
	consumer.Subscribe(topicList)
}

func TestConsumer_Subscribe_Topic_C(t *testing.T) {
	consumer := NewConsumer("tcp","127.0.0.1",9000,50)
	topicList := []string{"C"}
	consumer.Subscribe(topicList)
}