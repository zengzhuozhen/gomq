package client

import (
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/protocol"
	"testing"
)

func BenchmarkProducer_Publish(b *testing.B) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "A",
		Data: common.Message{
			MsgKey: "test",
			Body:   "hello world A ",
		},
	}
	for i := 0; i < b.N; i++ {
		producer.Publish(mess, protocol.AtMostOnce, protocol.NeedRetain)
	}
	producer.Close()

}

func BenchmarkProducer_Publish_Parallel(b *testing.B) {
	opts := defaultOpts()
	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "A",
		Data: common.Message{
			MsgKey: "test",
			Body:   "hello world A ",
		},
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			producer.Publish(mess, protocol.AtMostOnce, protocol.NeedRetain)
		}
	})
	producer.Close()
}

func TestProducer_Publish_Topic_A(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "A",
		Data: common.Message{
			MsgKey: "test",
			Body:   "hello world A ",
		},
	}
	producer.Publish(mess, protocol.AtLeastOnce, protocol.NeedRetain)
}

func TestProducer_Publish_Topic_B(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "B",
		Data: common.Message{
			MsgKey: "test",
			Body:   "hello world B ",
		},
	}
	producer.Publish(mess, protocol.AtLeastOnce, protocol.NeedRetain)
}

func TestProducer_Publish_Topic_C(t *testing.T) {
	opts := defaultOpts()

	producer := NewProducer(opts)
	mess := common.MessageUnit{
		Topic: "C",
		Data: common.Message{
			MsgKey: "test",
			Body:   "hello world C",
		},
	}
	producer.Publish(mess, protocol.ExactOnce, protocol.NotNeedRetain)
}
