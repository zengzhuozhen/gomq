package packet

import (
	"bytes"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
)

type SyncReqPacket struct {
	FixedHeader
	PacketIdentifier uint16
}

func NewSyncReqPacket(identity uint16) SyncReqPacket {
	return SyncReqPacket{
		FixedHeader:FixedHeader{TypeAndReserved: protocol.SYNCREQ},
		PacketIdentifier:identity,
	}
}

func (s *SyncReqPacket) Read(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	var err error
	s.PacketIdentifier, err = utils.DecodeUint16(r)
	return err
}

func (s *SyncReqPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	return nil
}

func (s *SyncReqPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint16(s.PacketIdentifier))
	packet := s.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
