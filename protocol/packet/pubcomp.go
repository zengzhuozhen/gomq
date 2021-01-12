package packet

import (
	"bytes"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
)

type PubCompPacket struct {
	FixedHeader
	PacketIdentifier uint16
}

func NewPubCompPacket(identity uint16) PubCompPacket {
	return PubCompPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.PUBCOMP)),
			RemainingLength: 2,
		},
		PacketIdentifier: identity,
	}
}

func (p *PubCompPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	var err error
	p.PacketIdentifier, err = utils.DecodeUint16(r)
	return err
}

func (c *PubCompPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}

func (p *PubCompPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(p.PacketIdentifier))
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
