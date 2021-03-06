package packet

import (
	"bytes"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
)

type PingReqPacket struct {
	FixedHeader
}



func NewPingReqPacket() PingReqPacket {
	return PingReqPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.PINGREQ)),
		},
	}
}

func (p *PingReqPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	return nil
}

func (c *PingReqPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}


func (p *PingReqPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
