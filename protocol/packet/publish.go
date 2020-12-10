package packet

import (
	"bytes"
	"fmt"
	"gomq/common"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type PublishPacket struct {
	FixedHeader
	TopicName        string
	PacketIdentifier uint16
	Payload          []byte
}

func NewPublishPacket(topic string, message common.Message, isFirst bool, QoS int, Retain int, identity uint16) PublishPacket {
	byte1 := utils.EncodePacketType(byte(protocol.PUBLISH)) // MQTT控制报文类型 (3)
	if !isFirst && QoS != protocol.AtMostOnce {
		byte1 += 8 // 重发标志 DUP
	}
	byte1 += byte(QoS << 1)    // 服务质量等级 QoS
	byte1 += byte(Retain) // 保留标志 RETAIN

	var identityLen int
	if QoS != 0 {
		identityLen = 2	// 占位
	}
	payLoad := message.Pack()
	return PublishPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: byte1,
			RemainingLength: len(utils.EncodeString(topic)) + identityLen + len(payLoad),
		},
		TopicName:        topic,
		PacketIdentifier: identity,
		Payload:          payLoad,
	}
}


func (p *PublishPacket) Read(r io.Reader, header FixedHeader) error {
	p.FixedHeader = header
	var payloadLength = header.RemainingLength
	var err error
	p.TopicName, err = utils.DecodeString(r)
	if err != nil {
		return nil
	}
	if header.QoS() > 0 {
		payloadLength -= len(p.TopicName) + 4		// has identity
		if p.PacketIdentifier,err = utils.DecodeUint16(r);err!= nil{
			return err
		}
	} else {
		payloadLength -= len(p.TopicName) + 2
	}
	if payloadLength < 0 {
		return fmt.Errorf("error unpacking publish, payload length < 0")
	}
	p.Payload = make([]byte, payloadLength)
	_, err = r.Read(p.Payload)
	return nil
}

func (c *PublishPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}


func (p *PublishPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(utils.EncodeString(p.TopicName))
	if p.PacketIdentifier != 0 {
		body.Write(utils.EncodeUint16(p.PacketIdentifier))
	}
	packet := p.FixedHeader.Pack()
	packet.Write(body.Bytes())
	packet.Write(p.Payload)
	_, err = w.Write(packet.Bytes())
	return err
}


func (p *PublishPacket) IsLegalQos(){
	if p.FixedHeader.QoS() > 3 {
		panic("非法的QoS标识，需要关闭连接")
	}
}
