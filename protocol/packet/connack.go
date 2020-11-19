package packet

import (
	"bytes"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

const (
	ConnectAccess = iota
	UnSupportProtocolVersion
	UnSupportClientIdentity
	UnAvailableService
	UserAndPassError
	UnAuthorization
)

type ConnAckPacket struct {
	FixedHeader
	ConnectAcknowledgeFlags byte
	ConnectReturnCode       byte
}

func NewConnectAckPack(code byte) ConnAckPacket {
	return ConnAckPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.CONNACK)),
			RemainingLength: 2,
		},
		ConnectAcknowledgeFlags: 1,
		ConnectReturnCode:       code,
	}
}

func (c *ConnAckPacket) Read(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	var err error
	c.ConnectAcknowledgeFlags,err  = utils.DecodeByte(r)
	c.ConnectReturnCode, err = utils.DecodeByte(r)
	return err
}

func (c *ConnAckPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}

func (c *ConnAckPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeByte(c.ConnectAcknowledgeFlags))
	body.Write(utils.EncodeByte(c.ConnectReturnCode))
	packet := c.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}

