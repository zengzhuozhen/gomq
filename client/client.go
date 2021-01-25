package client

import (
	"github.com/google/uuid"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/protocol"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"math"
	"net"
	"sync"
	"time"
)

type Option struct {
	Protocol string
	Address  string
	Timeout  int // timeout sec
	Username string
	Password string
}

type client struct {
	options      *Option
	optionsMu    sync.Mutex
	conn         net.Conn
	IdentityPool map[int]bool
	session      *common.ClientState
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
		session: &common.ClientState{
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
	payLoad := protocolPacket.NewConnectPayLoad(uuid.New().String(), "", "", c.options.Username, c.options.Password)
	connectPack := protocolPacket.NewConnectPacket(30, false, payLoad)
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

func (c *client) cleanSession() {
	c.session = &common.ClientState{
		Pub: make(map[uint16]interface{}),
		Rel: make(map[uint16]interface{}),
		Rec: make(map[uint16]interface{}),
	}
}

func (c *client) saveToSession(packet protocolPacket.ControlPacket) {
	switch packet.(type) {
	case *protocolPacket.PublishPacket:
		pubPacket := packet.(*protocolPacket.PublishPacket)
		c.session.Pub[pubPacket.PacketIdentifier] = *pubPacket
	case *protocolPacket.PubRelPacket:
		relPacket := packet.(*protocolPacket.PubRelPacket)
		c.session.Rel[relPacket.PacketIdentifier] = *relPacket
	case *protocolPacket.PubRecPacket:
		recPacket := packet.(*protocolPacket.PubRecPacket)
		c.session.Rec[recPacket.PacketIdentifier] = *recPacket
	}
}

func (c *client) deleteFromSession(packet protocolPacket.ControlPacket) {
	switch packet.(type) {
	case *protocolPacket.PubAckPacket:
		ackPacket := packet.(*protocolPacket.PubAckPacket)
		delete(c.session.Pub, ackPacket.PacketIdentifier)
	case *protocolPacket.PubCompPacket:
		compPacket := packet.(*protocolPacket.PubCompPacket)
		delete(c.session.Rec, compPacket.PacketIdentifier)
		delete(c.session.Rel, compPacket.PacketIdentifier)
	}
}
