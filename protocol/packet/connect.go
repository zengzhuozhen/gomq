package packet

import (
	"bytes"
	"fmt"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"io"
	"strings"
)

type ConnectPacket struct {
	FixedHeader
	ProtocolName  string // M,Q,T,T
	ProtocolLevel byte   // Level(4) //todo 协议等级
	ConnectFlags  byte   // todo
	KeepAlive     uint16 // 保持连接 Keep Alive MSB ,LSB
	payLoad       []byte
}

type ConnectFlags struct {
	CleanSession bool // todo 为true,则每次连接都是新的连接，为fasle，连接后需要处理旧的数据(qos相关)
	WillFlag     bool
	WillQos      uint8
	WillRetain   bool
	PasswordFlag bool
	UserNameFlag bool
}

type ConnectPacketPayLoad struct {
	ClientId    string // 客户端标识，非报文标识符
	WillTopic   string
	WillMessage string
	UserName    string
	Password    string
}

func NewConnectPayLoad(clientId, willTopic, willMessage, userName, password string) *ConnectPacketPayLoad {
	return &ConnectPacketPayLoad{
		ClientId:    clientId,
		WillTopic:   willTopic,
		WillMessage: willMessage,
		UserName:    userName,
		Password:    password,
	}
}

func NewConnectPacket(keepAlive uint16, cleanSession bool,WillQoS uint8, payLoad *ConnectPacketPayLoad) ConnectPacket {
	connectFlag, payLoadData := payLoad.encode(cleanSession,WillQoS)
	return ConnectPacket{
		FixedHeader: FixedHeader{
			TypeAndReserved: utils.EncodePacketType(byte(protocol.CONNECT)),
			RemainingLength: 10 + len(payLoadData),
			//剩余长度等于可变报头的长度（10字节）加上有效载荷的长度
		},
		// 可变包头
		ProtocolName:  "MQTT",
		ProtocolLevel: byte(4),
		ConnectFlags:  connectFlag.encode(),
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

// 如果发现不支持的协议级别，服务端必须给发送一个返回码为0x01（不支持的协议级别）的CONNACK报文响应CONNECT报文，然后断开客户端的连接
func (c *ConnectPacket) IsSuitableProtocolLevel() bool {
	return c.ProtocolLevel == 4
}

// 服务端必须判断 reserved 是否为0，不为0就要断开客户端连接
func (c *ConnectPacket) IsReserved() bool {
	return c.ConnectFlags%2 == 0
}

func (c *ConnectPacketPayLoad) IsLegalClientId() bool {
	mode := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, i := range c.ClientId {
		if !strings.Contains(mode, string(i)) {
			return false
		}
	}
	return true
}

func (c *ConnectPacketPayLoad) IsAuthorizedClient() bool {
	//todo 验证客户端是否之前授权了
	return true
}

func (c *ConnectPacketPayLoad) IsCorrectSecret(getSecretFunc func(string) string) bool {
	if c.UserName == ""{
		return true
	}
	return  getSecretFunc(c.UserName) == c.Password
}

// ProvisionConnectFlags Provision Connection Flags
func (c *ConnectPacket) ProvisionConnectFlags() *ConnectFlags {
	connectFlags := new(ConnectFlags)
	connectFlags.decode(c.ConnectFlags)
	return connectFlags
}

// ProvisionConnectFlagsAndPayLoad 根据 connectPacket包提取 connectFlags 和 payLoad 结构
func (c *ConnectPacket) ProvisionConnectFlagsAndPayLoad() (*ConnectFlags, *ConnectPacketPayLoad) {
	connectFlags := c.ProvisionConnectFlags()
	payLoad := new(ConnectPacketPayLoad)
	payLoad.decode(c.payLoad, *connectFlags)
	return connectFlags, payLoad
}

func (c *ConnectPacket) Visit(fn VisitorFunc) error {
	return fn(c)
}

// 保留位 Reserved  			位置：连接标志字节的第0位
// 清理会话 Clean Session 	位置：连接标志字节的第1位      +2
// 遗嘱标志 Will Flag 		位置：连接标志的第2位。	    +4
// 遗嘱QoS Will QoS			位置：连接标志的第4和第3位。   +8 +16
// 遗嘱保留 Will Retain	   	位置：连接标志的第5位。	   +32
// 密码标志 Password Flag    位置：连接标志的第6位。	   +64
// 用户名标志 User Name Flag 位置：连接标志的第7位。	   	   +128

func (c *ConnectFlags) encode() byte {
	res := byte(0)
	if c.CleanSession {
		res += 1 << 1
	}
	if c.WillFlag {
		res += 1 << 2
	}
	 if c.WillQos == 1{
		res += 1<<3
	}else if c.WillQos == 2{
		res += 1<<4
	}
	if c.WillRetain {
		res += 1 << 5
	}
	if c.PasswordFlag {
		res += 1 << 6
	}
	if c.UserNameFlag {
		res += 1 << 7
	}
	return res
}

func (c *ConnectFlags) decode(b byte) () {
	if b >= 128 {
		b -= 128
		c.UserNameFlag = true
	}
	if b >= 64 {
		b -= 64
		c.PasswordFlag = true
	}
	if b >= 32 {
		b -= 32
		c.WillRetain = true
	}
	if b >= 16 {
		b -= 16
		c.WillQos = 2
	}
	if b >= 8 {
		b -= 8
		c.WillQos = 1
	}
	if b >= 4 {
		b -= 4
		c.WillFlag = true
	}
	if b >= 2 {
		b -= 2
		c.CleanSession = true
	}
	return
}

func (payLoad ConnectPacketPayLoad) encode(cleanSession bool,WillQoS uint8) (*ConnectFlags, []byte) {
	var payLoadData []byte
	connectFlag := new(ConnectFlags)
	connectFlag.CleanSession = cleanSession

	temp := utils.EncodeString(payLoad.ClientId)
	payLoadData = append(payLoadData, temp...)

	if payLoad.WillTopic != "" && payLoad.WillMessage != "" {
		connectFlag.WillFlag = true
		connectFlag.WillQos = WillQoS
		temp = utils.EncodeString(payLoad.WillTopic)
		payLoadData = append(payLoadData, temp...)
		temp = utils.EncodeString(payLoad.WillMessage)
		payLoadData = append(payLoadData, temp...)
	}
	if payLoad.UserName != "" {
		connectFlag.UserNameFlag = true
		temp = utils.EncodeString(payLoad.UserName)
		payLoadData = append(payLoadData, temp...)
	}
	if payLoad.Password != "" {
		connectFlag.PasswordFlag = true
		temp = utils.EncodeString(payLoad.Password)
		payLoadData = append(payLoadData, temp...)
	}
	return connectFlag, payLoadData
}

func (payLoad ConnectPacketPayLoad) decode(payLoadData []byte, flags ConnectFlags) {
	var err error
	buffer := new(bytes.Buffer)
	buffer.Write(payLoadData)
	payLoad.ClientId, err = utils.DecodeString(buffer)
	if flags.WillFlag {
		payLoad.WillTopic, err = utils.DecodeString(buffer)
		payLoad.WillMessage, err = utils.DecodeString(buffer)
	}
	if flags.UserNameFlag {
		payLoad.UserName, err = utils.DecodeString(buffer)
	}
	if flags.PasswordFlag {
		payLoad.Password, err = utils.DecodeString(buffer)
	}
	if err != nil {
		panic("decode payload failed")
	}
}
