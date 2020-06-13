package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type PubRecPacket struct {
	FixedHeader
	PacketIdentifier uint16
}

func NewPubRecPacket(identity uint16) PubRecPacket {
	return PubRecPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: EncodePacketType(byte(protocol.PUBREC)),
			RemainingLength: 2,
		},
		PacketIdentifier: identity,
	}
}

func (p *PubRecPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	var err error
	p.PacketIdentifier, err = utils.DecodeUint16(r)
	return err
}

func (p *PubRecPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(p.PacketIdentifier))
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
