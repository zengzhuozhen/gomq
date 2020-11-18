package client

import (
	"gomq/common"
)

type Member struct {
	opts   *Option
	client *client
}

func NewMember(opts *Option) *Member {
	client := NewClient(opts).(*client)
	return &Member{client: client,opts:opts}
}

func (m *Member) StartConsume(queue *common.Queue) error{
	consumer := NewConsumer(&Option{
		Protocol: m.opts.Protocol,
		Host:     m.opts.Host,
		Port:     m.opts.Port,
		Timeout:  m.opts.Timeout,
	})
	for msgUnit := range consumer.Subscribe([]string{"*"}) {
		queue.Push(*msgUnit)
	}
	return nil
}

