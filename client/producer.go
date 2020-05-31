package client

import (
	"encoding/json"
	"gomq/common"
)

type Producer struct {
	bc *BasicClient
}

func NewProducer(protocol, host string, port, timeout int) *Producer {
	return &Producer{bc: &BasicClient{
			protocol:protocol,
			host:host,
			port:port,
			timeout:timeout,
	}}
}

func (p *Producer) Publish(mess common.Message) {
	conn := p.bc.Connect()
	netPacket := common.NewPacket(common.P,mess)
	data, err := json.Marshal(netPacket)
	if err != nil {
		return
	}
	conn.Write(data)
}
