package packet

import (
	"bytes"
	"fmt"
	"gomq/protocol"
	"gomq/protocol/utils"
	"io"
)

type ConnectPacket struct {
	FixedHeader
	ProtocolName  string // M,Q,T,T
	ProtocolLevel byte   // Level(4) //todo 协议等级
	ConnectFlags  byte   // todo
	KeepAlive     uint16 // 保持连接 Keep Alive MSB ,LSB
	payLoad       []byte
}
type ConnectPacketPayLoad struct {
	identifier  string
	willTopic   string
	willMessage string
	userName    string
	password    string
}

func NewConnectPack(
	keepAlive uint16,
	cleanSession bool,
	payLoadStruct ConnectPacketPayLoad,
) ConnectPacket {
	var ConnectPayLoadLength int
	var payLoadData []byte
	var WillFlag, WillQoS, WillRetain, UserNameFlag, PasswordFlag bool

	temp := utils.EncodeString(payLoadStruct.identifier)
	payLoadData = append(payLoadData, temp...)
	ConnectPayLoadLength += len(temp)

	if payLoadStruct.willTopic != "" && payLoadStruct.willMessage != "" {
		WillFlag = true
		temp = utils.EncodeString(payLoadStruct.willTopic)
		payLoadData = append(payLoadData, temp...)
		ConnectPayLoadLength += len(temp)
		temp = utils.EncodeString(payLoadStruct.willMessage)
		payLoadData = append(payLoadData, temp...)
		ConnectPayLoadLength += len(temp)
	}
	if payLoadStruct.userName != "" {
		UserNameFlag = true
		temp = utils.EncodeString(payLoadStruct.userName)
		payLoadData = append(payLoadData, temp...)
		ConnectPayLoadLength += len(temp)
	}
	if payLoadStruct.password != "" {
		PasswordFlag = true
		temp = utils.EncodeString(payLoadStruct.password)
		payLoadData = append(payLoadData, temp...)
		ConnectPayLoadLength += len(temp)
	}

	return ConnectPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: EncodePacketType(byte(protocol.CONNECT)),
			RemainingLength: 10 + ConnectPayLoadLength,
			//剩余长度等于可变报头的长度（10字节）加上有效载荷的长度
		},
		// 可变包头
		ProtocolName:  "MQTT",
		ProtocolLevel: byte(4),
		ConnectFlags:  EncodeConnectFlag(cleanSession, WillFlag, WillQoS, WillRetain, UserNameFlag, PasswordFlag),
		KeepAlive:     keepAlive,
		// 有效载荷
		payLoad: payLoadData,
	}
}

func (c *ConnectPacket) Read(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	var err error
	c.ProtocolName, _ = utils.DecodeString(r)
	c.ProtocolLevel, _ = utils.DecodeByte(r)
	c.ConnectFlags, _ = utils.DecodeByte(r)
	c.KeepAlive, _ = utils.DecodeUint16(r)
	var payloadLength = header.RemainingLength - 10
	c.payLoad = make([]byte, payloadLength)

	if payloadLength < 0 {
		return fmt.Errorf("error unpacking publish, payload length < 0")
	}
	c.payLoad = make([]byte, payloadLength)
	_, err = r.Read(c.payLoad)
	return err
}

func (c *ConnectPacket) ReadHeadOnly(r io.Reader, header FixedHeader) error {
	c.FixedHeader = header
	return nil
}

func (c *ConnectPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error
	body.Write(utils.EncodeString(c.ProtocolName))
	body.Write(utils.EncodeByte(c.ProtocolLevel))
	body.Write(utils.EncodeByte(c.ConnectFlags))
	body.Write(utils.EncodeUint16(c.KeepAlive))
	body.Write(c.payLoad)
	packet := c.FixedHeader.Pack()
	packet.Write(body.Bytes())
	_, err = w.Write(packet.Bytes())
	return err
}

// 服务端必须判断 reserved 是否为0，不为0就要断开客户端连接
func (c *ConnectPacket) IsReserved() bool {
	if c.ConnectFlags%2 == 0 {
		return true
	}
	return false
}

func (c *ConnectPacket) IsLegalIdentifier() bool {
	_ = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// todo identity 合法性验证
	return true
}

func (c *ConnectPacket) IsAuthorizedClient() bool {
	//todo 验证客户端是否之前授权了
	return true
}

func (c *ConnectPacket) IsCorrectSecret() bool {
	//todo key 和 secret 验证
	return true
}

// todo 清理会话 Clean Session 	位置：连接标志字节的第1位     +1
// todo 遗嘱标志 Will Flag 		位置：连接标志的第2位。	   +2
// todo 遗嘱QoS Will QoS			位置：连接标志的第4和第3位。   +4 +8
// todo 遗嘱保留 Will Retain	   	位置：连接标志的第5位。	   +16
// todo 密码标志 password Flag   位置：连接标志的第6位。	   +32
// todo 用户名标志 User Name Flag 位置：连接标志的第7位。	   +64

func EncodeConnectFlag(CleanSession bool, WillFlag bool, WillQos bool, WillRetain bool, UserName bool, Password bool, ) byte {
	res := byte(0)
	if CleanSession {
		res += 1
	}
	if WillFlag {
		res += 1 << 1
	}
	if WillQos {
		res += 1<<2 + 1<<3
	}
	if WillRetain {
		res += 1 << 4
	}
	if Password {
		res += 1 << 5
	}
	if UserName {
		res += 1 << 6
	}
	return res
}

func DecodeConnectFlag(b byte) (CleanSession bool, WillFlag bool, WillQos bool, WillRetain bool, UserName bool, Password bool) {
	if b >= 64 {
		b -= 64
		UserName = true
	}
	if b >= 32 {
		b -= 32
		Password = true
	}
	if b >= 16 {
		b -= 16
		WillRetain = true
	}
	if b >= 12 {
		b -= 12
		WillQos = true
	}
	if b >= 2 {
		b -= 2
		WillFlag = true
	}
	if b >= 1 {
		b -= 1
		CleanSession = true
	}
	return
}

func EncodePacketType(i byte) byte {
	var res byte
	if i >= 8 {
		i -= 8
		res += 128
	}
	if i >= 4 {
		i -= 4
		res += 64
	}
	if i >= 2 {
		i -= 2
		res += 32
	}
	if i >= 1 {
		i -= 1
		res += 16
	}
	return res
}

func DecodePacketType(byte1 byte) byte {
	var res byte
	if byte1 >= 128 {
		byte1 -= 128
		res += 8
	}
	if byte1 >= 64 {
		byte1 -= 64
		res += 4
	}
	if byte1 >= 32 {
		byte1 -= 32
		res += 2
	}
	if byte1 >= 16 {
		byte1 -= 16
		res += 1
	}
	return res
}
