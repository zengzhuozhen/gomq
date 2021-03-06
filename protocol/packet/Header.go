package packet

import (
	"bytes"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
)

type FixedHeader struct {
	TypeAndReserved byte // MQTT报文类型 + Reserved 保留位
	RemainingLength int  // 剩余长度 ,最大4个字节
}

func (f *FixedHeader) Read(r io.Reader) error {
	b := make([]byte, 1)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	return f.UnPack(b[0], r) //
}

func (f *FixedHeader) Pack() bytes.Buffer {
	var header bytes.Buffer
	header.WriteByte(f.TypeAndReserved)
	header.Write(utils.EncodeRemainingLengthAlg(f.RemainingLength))
	return header
}

func (f *FixedHeader) UnPack(byte1 byte, r io.Reader) (err error) {
	f.TypeAndReserved = byte1
	f.RemainingLength, err = utils.DecodeRemainingLengthAlg(r)
	return err
}

// QoS 服务质量等级
func (f *FixedHeader) QoS() byte {
	return (f.TypeAndReserved >> 1) & 0x03
}

// Dup 重发标识
func (f *FixedHeader) Dup() bool {
	return (f.TypeAndReserved>>3)&0x01 > 0
}

// Retain 保留标志
func (f *FixedHeader) Retain() bool {
	return f.TypeAndReserved&0x01 > 0
}
