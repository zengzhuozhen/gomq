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
	byte1 := EncodePacketType(byte(protocol.PUBLISH)) // MQTT控制报文类型 (3)
	if !isFirst && QoS == protocol.ExactOnce {
		byte1 += 8 // 重发标志 DUP
	}
	byte1 += byte(QoS << 1)    // 服务质量等级 QoS
	byte1 += byte(Retain) // 保留标志 RETAIN

	var identityLen int
	if QoS == 1 || QoS == 2 {
		identityLen = 2	// 占位
	}
	payLoad := message.Pack()
	fmt.Println(payLoad)
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

// 如果客户端发给服务端的PUBLISH报文的保留（RETAIN）标志被设置为1，服务端必须存储这个应用消息和它的服务质量等级（QoS）
//，以便它可以被分发给未来的主题名匹配的订阅者 [MQTT-3.3.1-5]。
//一个新的订阅建立时，对每个匹配的主题名，如果存在最近保留的消息，它必须被发送给这个订阅者 [MQTT-3.3.1-6]。
//如果服务端收到一条保留（RETAIN）标志为1的QoS 0消息，它必须丢弃之前为那个主题保留的任何消息。
//它应该将这个新的QoS 0消息当作那个主题的新保留消息，但是任何时候都可以选择丢弃它 — 如果这种情况发生了，那个主题将没有保留消息 [MQTT-3.3.1-7]。
func HandleRetain() {
	// todo
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
		fmt.Println(p.PacketIdentifier)
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
