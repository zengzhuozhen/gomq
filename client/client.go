package client

import (
	"github.com/google/uuid"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/protocol"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"math"
	"net"
	"sync"
	"time"
)

type Option struct {
	Protocol     string
	Address      string
	Timeout      int // timeout sec
	Username     string
	Password     string
	WillTopic    string
	WillMessage  string
	WillQoS      uint8
	CleanSession bool
	KeepAlive    uint16
}

type client struct {
	options      *Option
	optionsMu    sync.Mutex
	conn         net.Conn
	IdentityPool map[int]bool
	session      *PacketStateSession
}

func newClient(opt *Option) *client {
	identityPool := make(map[int]bool)
	for i := 1; i <= math.MaxUint16; i++ { // 非零的16位报文标识符
		identityPool[i] = true
	}
	return &client{
		options:      opt,
		optionsMu:    sync.Mutex{},
		IdentityPool: identityPool,
		session: &PacketStateSession{
			Pub: make(map[uint16]interface{}),
			Rel: make(map[uint16]interface{}),
			Rec: make(map[uint16]interface{}),
		},
	}
}

func (c *client) connect() error {
	conn, err := net.DialTimeout(c.options.Protocol, c.options.Address, time.Duration(c.options.Timeout)*time.Second)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	c.conn = conn
	payLoad := protocolPacket.NewConnectPayLoad(uuid.New().String(), c.options.WillTopic, c.options.WillMessage, c.options.Username, c.options.Password)
	connectPack := protocolPacket.NewConnectPacket(c.options.KeepAlive, c.options.CleanSession, c.options.WillQoS, payLoad)
	connectPack.Write(conn)

	var fh protocolPacket.FixedHeader
	var connAckPacket protocolPacket.ConnAckPacket

	if err := fh.Read(conn); err != nil {
		log.Errorf("接收connAck包头内容错误", err)
		return err
	}
	err = connAckPacket.Read(conn, fh)
	if err != nil {
		log.Errorf("读取connAck数据包内容错误", err)
		return err
	}

	if connAckPacket.ConnectReturnCode != protocol.ConnectAccess {
		log.Errorf("服务器拒绝连接，code:", connAckPacket.ConnectReturnCode)
		conn.Close()
	}
	log.Infof("接收返回报文 connectAck, return code:", connAckPacket.ConnectReturnCode)
	return nil
}

func (c *client) disConnect() {
	disConnectPack := protocolPacket.NewDisConnectPacketPacket()
	_ = disConnectPack.Write(c.conn)
	_ = c.conn.Close()
}

func (c *client) getAvailableIdentity() uint16 {
	for k, v := range c.IdentityPool {
		if v == true {
			c.IdentityPool[k] = false
			return uint16(k)
		}
	}
	return 0
}

// save
// 保存已发送的publish,已收到的pubRec，已发送的pubRel
func (c *PacketStateSession) save(packet protocolPacket.ControlPacket) {
	switch packet.(type) {
	case *protocolPacket.PublishPacket:
		pubPacket := packet.(*protocolPacket.PublishPacket)
		c.Pub[pubPacket.PacketIdentifier] = *pubPacket
	case *protocolPacket.PubRecPacket:
		recPacket := packet.(*protocolPacket.PubRecPacket)
		c.Rec[recPacket.PacketIdentifier] = *recPacket
	case *protocolPacket.PubRelPacket:
		relPacket := packet.(*protocolPacket.PubRelPacket)
		c.Rel[relPacket.PacketIdentifier] = *relPacket
	}
}

// remove
// 收到Ack，删除Pub
// 收到Rec，删除Pub
// 收到Rel，删除Rec
// 收到Comp，删除Rel
func (c *PacketStateSession) remove(packet protocolPacket.ControlPacket) {
	switch packet.(type) {
	case *protocolPacket.PubAckPacket:
		ackPacket := packet.(*protocolPacket.PubAckPacket)
		delete(c.Pub, ackPacket.PacketIdentifier)
	case *protocolPacket.PubRecPacket:
		recPacket := packet.(*protocolPacket.PubRecPacket)
		delete(c.Pub, recPacket.PacketIdentifier)
	case *protocolPacket.PubRelPacket:
		relPacket := packet.(*protocolPacket.PubRelPacket)
		delete(c.Rec, relPacket.PacketIdentifier)
	case *protocolPacket.PubCompPacket:
		compPacket := packet.(*protocolPacket.PubCompPacket)
		delete(c.Rel, compPacket.PacketIdentifier)
	}
}

// PacketStateSession 保存通讯过程中包的中间状态，以便崩溃重启后重发消息，保证消息不丢失
// 只在客户端实现，通过客户端驱动服务端进行重发，服务端无状态
type PacketStateSession struct {
	// 已经发送给服务端，但是还没有完成确认的QoS 1和QoS 2级别的消息
	Pub map[uint16]interface{}
	// 已从服务端接收，但是还没有发送Rel的QoS 2级别信息
	Rec map[uint16]interface{}
	// 已经发送给服务端，但是还没有完成确认的QoS 2级别的消息
	Rel map[uint16]interface{}
}
