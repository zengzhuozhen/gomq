package client

import (
	"gomq/common"
	"testing"
)

func BenchmarkProducer_Publish(b *testing.B) {
	opts := defaultOpts()

	for i := 0; i < b.N; i++ {
		producer := NewProducer(opts)
		mess := common.MessageUnit{
			Topic: "A",
			Data: common.Message{
				MsgId:  1,
				MsgKey: "test",
				Body:   "hello world A ",
			},
		}
		producer.Publish(mess, 0, 0)
	}
}

func BenchmarkProducer_Publish_Parallel(b *testing.B) {
	opts := defaultOpts()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			producer := NewProducer(opts)
			mess := common.MessageUnit{
				Topic: "A",
				Data: common.Message{
					MsgId:  1,
					MsgKey: "test",
					Body:   "hello world A ",
				},
			}
			producer.Publish(mess, 0, 0)
		}
	})
}

func TestProducer_Publish_Topic_A(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "A",
		Data: common.Message{
			MsgId:  1,
			MsgKey: "test",
			Body:   "hello world A ",
		},
	}
	producer.Publish(mess, 1, 1)
}

func TestProducer_Publish_Topic_B(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "B",
		Data: common.Message{
			MsgId:  2,
			MsgKey: "test",
			Body:   "hello world B ",
		},
	}
	producer.Publish(mess, 1, 1)
}

func TestProducer_Publish_Topic_C(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "C",
		Data: common.Message{
			MsgId:  3,
			MsgKey: "test",
			Body:   "hello world C",
		},
	}
	producer.Publish(mess, 2, 0)
}
