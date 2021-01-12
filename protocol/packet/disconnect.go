package packet

import (
	"bytes"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
)

type DisConnectPacket struct {
	FixedHeader
}



func NewDisConnectPacketPacket() DisConnectPacket {
	return DisConnectPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.DISCONNECT)),
		},
	}
}

func (d *DisConnectPacket) Read(r io.Reader, header FixedHeader) error {
	d.FixedHeader = header
	return nil
}

func (c *DisConnectPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}


func (d *DisConnectPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	packet := d.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}
