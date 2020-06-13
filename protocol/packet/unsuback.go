package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type UnSubAckPacket struct {
	FixedHeader
	PacketIdentifier uint16
}



func NewUnSubAckPacket(identity uint16) UnSubAckPacket {
	return UnSubAckPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: EncodePacketType(byte(protocol.UNSUBACK)),
			RemainingLength: 2,
		},
		PacketIdentifier: identity,
	}
}

func (s *UnSubAckPacket) Read(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	var err error
	s.PacketIdentifier,err = utils.DecodeUint16(r);
	return err
}

func (s *UnSubAckPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(s.PacketIdentifier))
	packet := s.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
