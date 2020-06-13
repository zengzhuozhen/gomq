package packet

import (
	"bytes"
	"gomq/protocol"
	"io"
)

type PingReqPacket struct {
	FixedHeader
}



func NewPingReqPacket() PingReqPacket {
	return PingReqPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: EncodePacketType(byte(protocol.PINGREQ)),
		},
	}
}

func (p *PingReqPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	var err error
	return err
}

func (p *PingReqPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
