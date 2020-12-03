package client

import (
	"fmt"
	"gomq/protocol"
	"gomq/protocol/packet"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

type Client interface {
	Connect() error
	GetAvailableIdentity() uint16
}

type Option struct {
	Protocol string
	Address  string
	Timeout  int // timeout sec
}

type client struct {
	options      *Option
	optionsMu    sync.Mutex
	conn         net.Conn
	IdentityPool map[int]bool
}

func NewClient(opt *Option) Client {
	identityPool := make(map[int]bool)
	for i := 1; i <= math.MaxUint16; i++ { // 非零的16位报文标识符
		identityPool[i] = true
	}
	return &client{
		options:      opt,
		optionsMu:    sync.Mutex{},
		IdentityPool: identityPool,
	}
}

func (c *client) Connect() error {
	conn, err := net.DialTimeout(c.options.Protocol, c.options.Address, time.Duration(c.options.Timeout)*time.Second)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	c.conn = conn
	payLoad := new(packet.ConnectPacketPayLoad)
	//todo cleanSession 暂时设置为 false
	connectPack := packet.NewConnectPack(30, false, *payLoad)
	connectPack.Write(conn)

	var fh packet.FixedHeader
	var connAckPacket packet.ConnAckPacket

	if err := fh.Read(conn); err != nil {
		fmt.Println("接收connAck包头内容错误", err)
		return err
	}
	err = connAckPacket.Read(conn, fh)
	if err != nil {
		log.Fatal("读取connAck数据包内容错误", err)
		return err
	}

	if connAckPacket.ConnectReturnCode != protocol.ConnectAccess {
		log.Fatal("服务器拒绝连接，code:", connAckPacket.ConnectReturnCode)
		conn.Close()
	}
	fmt.Println("接收返回报文 connectAck, return code:", connAckPacket.ConnectReturnCode)

	return nil
}

func (c *client) GetAvailableIdentity() uint16 {
	for k, v := range c.IdentityPool {
		if v == true {
			c.IdentityPool[k] = false
			return uint16(k)
		}
	}
	return 0
}
