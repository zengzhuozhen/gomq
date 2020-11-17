package client

import "gomq/common"

type Member struct {
	opts   *Option
	client *client
}

func NewMember(opts *Option) *Member {
	client := NewClient(opts).(*client)
	return &Member{client: client}
}

func (m *Member) StartConsume(msgChan <-chan *common.Message) {
	consumer := NewConsumer(&Option{
		Protocol: m.opts.Protocol,
		Host:     m.opts.Host,
		Port:     m.opts.Port,
		Timeout:  m.opts.Timeout,
	})
	msgChan = consumer.Subscribe([]string{"*"})
}
