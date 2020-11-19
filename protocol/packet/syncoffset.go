package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type SyncOffsetPacket struct {
	FixedHeader
	Offset uint64
	PacketIdentifier uint16
}

func NewSyncOffsetPacket(offset uint64,identity uint16) SyncOffsetPacket {
	return SyncOffsetPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: protocol.SYNCOFFSET,
			RemainingLength: 10, // 8个字节的offset + 2字节的identity
		},
		Offset: offset,
		PacketIdentifier: identity,
	}
}

func (s *SyncOffsetPacket) Read(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	var err error
	s.Offset, err = utils.DecodeUint64(r)
	s.PacketIdentifier,err = utils.DecodeUint16(r)
	return err
}

func (s *SyncOffsetPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	s.FixedHeader = header
	return nil
}

func (s *SyncOffsetPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeUint64(s.Offset))
	body.Write(utils.EncodeUint16(s.PacketIdentifier))
	packet := s.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
