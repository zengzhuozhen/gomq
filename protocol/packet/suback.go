package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type SubAckPacket struct {
	FixedHeader
	PacketIdentifier uint16
	Payload          []byte
}

func NewSubAckPacket(Identifier uint16,QoSList []byte) SubAckPacket {
	var payLoad []byte
	for _,v := range QoSList{
		payLoad = append(payLoad,v)
	}
	return SubAckPacket{
		FixedHeader:      FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.SUBACK)), // MQTT控制报文类型 (3)
			RemainingLength: 2 + len(payLoad),
		},
		PacketIdentifier: Identifier,
		Payload:          payLoad,
	}
}

func (s *SubAckPacket) Read(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	var payloadLength = header.RemainingLength -2

	var err error
	s.PacketIdentifier, err = utils.DecodeUint16(r)
	s.Payload = make([]byte, payloadLength)
	_, err = r.Read(s.Payload)
	return err
}

func (s *SubAckPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	return nil
}

func (s *SubAckPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(s.PacketIdentifier))
	body.Write(s.Payload)
	packet := s.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}


