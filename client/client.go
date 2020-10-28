package client

import (
	"fmt"
	"gomq/protocol/packet"
	"log"
	"math"
	"net"
	"strconv"
	"sync"
	"time"
)

type Client interface {
	Connect() error
}

type client struct {
	options      *Option
	optionsMu    sync.Mutex
	conn         net.Conn
	IdentityPool map[int]bool
}

func (c *client) Connect() error {
	conn, err := net.DialTimeout(c.options.Protocol, c.options.Host+":"+strconv.Itoa(c.options.Port), time.Duration(c.options.Timeout)*time.Second)
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
	}
	err = connAckPacket.Read(conn, fh)
	if err != nil {
		log.Fatal("读取connAck数据包内容错误", err)
		return err
	}

	if connAckPacket.ConnectReturnCode != packet.ConnectAccess {
		log.Fatal("服务器拒绝连接，code:", connAckPacket.ConnectReturnCode)
		conn.Close()
	}
	fmt.Println("接收返回报文 connectAck, return code:", connAckPacket.ConnectReturnCode)

	c.IdentityPool = make(map[int]bool)
	for i := 1; i <= math.MaxUint16; i++ { // 非零的16位报文标识符
		c.IdentityPool[i] = true
	}
	return nil
}

type Option struct {
	Protocol string
	Host     string
	Port     int
	Timeout  int // timeout sec
}
