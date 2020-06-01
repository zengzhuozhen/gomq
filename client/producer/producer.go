package producer

import (
	"encoding/json"
	"gomq/client"
	"gomq/common"
)

type Producer struct {
	bc *client.BasicClient
}

func NewProducer(protocol, host string, port, timeout int) *Producer {
	return &Producer{bc: &client.BasicClient{
			Protocol:protocol,
			Host:host,
			Port:port,
			Timeout:timeout,
	}}
}

func (p *Producer) Publish(topic string ,mess common.Message) {
	conn := p.bc.Connect()
	netPacket := common.Packet{
		Flag:          common.P,
		Message:       mess,
		Topic:         topic,
	}
	data, err := json.Marshal(&netPacket)
	if err != nil {
		return
	}
	conn.Write(data)
}
