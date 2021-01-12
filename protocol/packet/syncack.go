package packet

import (
	"bytes"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
)

type SyncAckPacket struct {
	FixedHeader
	PacketIdentifier uint16
}

func NewSyncAckPacket(identity uint16) SyncAckPacket {
	return SyncAckPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: protocol.SYNCACK,
			RemainingLength: 2,
		},
		PacketIdentifier: identity,
	}
}

func (s *SyncAckPacket) Read(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	var err error
	s.PacketIdentifier, err = utils.DecodeUint16(r)
	return err
}

func (s *SyncAckPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	return nil
}

func (s *SyncAckPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(s.PacketIdentifier))
	packet := s.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
