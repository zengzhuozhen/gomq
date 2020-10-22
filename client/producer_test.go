package client

import (
	"gomq/common"
	"testing"
)

func BenchmarkProducer_Publish(b *testing.B) {
	opts := defaultOpts()

	for i := 0; i < b.N; i++ {
		producer := NewProducer(opts)
		mess := common.Message{
			MsgId:  1,
			MsgKey: "test",
			Body:   "hello world A ",
		}
		producer.Publish("A", mess, 0)
	}
}

func BenchmarkProducer_Publish_Parallel(b *testing.B) {
	opts := defaultOpts()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			producer := NewProducer(opts)
			mess := common.Message{
				MsgId:  1,
				MsgKey: "test",
				Body:   "hello world A ",
			}
			producer.Publish("A", mess, 0)
		}
	})
}


func TestProducer_Publish_Topic_A(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.Message{
		MsgId:  1,
		MsgKey: "test",
		Body:   "hello world A ",
	}
	producer.Publish("A", mess, 0)
}

func TestProducer_Publish_Topic_B(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.Message{
		MsgId:  2,
		MsgKey: "test",
		Body:   "hello world B",
	}
	producer.Publish("B", mess, 1)
}

func TestProducer_Publish_Topic_C(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.Message{
		MsgId:  3,
		MsgKey: "test",
		Body:   "hello world C",
	}
	producer.Publish("C", mess, 2)
}

