package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type UnSubscribePacket struct {
	FixedHeader
	PacketIdentifier uint16
	Payload          []byte
}

func NewUnSubscribePacket(Identifier uint16,topicList []string) UnSubscribePacket {
	var payLoad []byte
	for _,v := range topicList{
		payLoad = append(payLoad,utils.EncodeString(v)...)
	}
	return UnSubscribePacket{
		FixedHeader:      FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.UNSUBSCRIBE)), // MQTT控制报文类型 (3)
			RemainingLength: 2 + len(payLoad),
		},
		PacketIdentifier: Identifier,
		Payload:          payLoad,
	}
}

func (s *UnSubscribePacket) Read(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	var payloadLength = header.RemainingLength -2
	var err error
	s.PacketIdentifier, err = utils.DecodeUint16(r)
	s.Payload = make([]byte, payloadLength)
	_, err = r.Read(s.Payload)
	return err
}

func (c *UnSubscribePacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}

func (s *UnSubscribePacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(s.PacketIdentifier))
	body.Write(s.Payload)
	packet := s.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}


