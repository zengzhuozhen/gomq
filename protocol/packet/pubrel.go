package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type PubRelPacket struct {
	FixedHeader
	PacketIdentifier uint16
}

func NewPubRelPacket(identity uint16) PubRelPacket {
	return PubRelPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: EncodePacketType(byte(protocol.PUBREL)),
			RemainingLength: 2,
		},
		PacketIdentifier: identity,
	}
}

func (p *PubRelPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	var err error
	p.PacketIdentifier, err = utils.DecodeUint16(r)
	return err
}

func (c *PubRelPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}

func (p *PubRelPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(p.PacketIdentifier))
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
