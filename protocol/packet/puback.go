package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type PubAckPacket struct {
	FixedHeader
	PacketIdentifier uint16
}



func NewPubAckPacket(identity uint16) PubAckPacket {
	return PubAckPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: EncodePacketType(byte(protocol.PUBACK)),
			RemainingLength: 2,
		},
		PacketIdentifier: identity,
	}
}

func (p *PubAckPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	var err error
	p.PacketIdentifier,err = utils.DecodeUint16(r);
	return err
}

func (p *PubAckPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(p.PacketIdentifier))
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
