package packet

import (
	"bytes"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
)

type PingRespPacket struct {
	FixedHeader
}



func NewPingRespPacket() PingRespPacket {
	return PingRespPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.PINGRESP)),
		},
	}
}

func (p *PingRespPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	return nil
}

func (c *PingRespPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}

func (p *PingRespPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}

