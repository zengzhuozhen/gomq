package client

import (
	"fmt"
	"gomq/protocol/packet"
	"log"
	"math"
	"net"
	"strconv"
	"time"
)

type BasicClient struct {
	Protocol     string
	Host         string
	Port         int
	Timeout      int // timeout sec
	IdentityPool map[int]bool
}

func (b *BasicClient) Connect() (net.Conn, error) {
	conn, err := net.DialTimeout(b.Protocol, b.Host+":"+strconv.Itoa(b.Port), time.Duration(b.Timeout)*time.Second)
	if err != nil {
		log.Fatal(err.Error())
		return nil, err
	}
	payLoad := new(packet.ConnectPacketPayLoad)
	//todo cleanSession 暂时设置为 false
	connectPack := packet.NewConnectPack( 30,false,*payLoad)
	connectPack.Write(conn)

	var fh packet.FixedHeader
	var connAckPacket packet.ConnAckPacket
	if err := fh.Read(conn);err !=nil{
		fmt.Println("接收connAck包头内容错误",err)
	}
	err = connAckPacket.Read(conn, fh)
	if err != nil {
		log.Fatal("读取connAck数据包内容错误",err)
		return nil, err
	}

	if connAckPacket.ConnectReturnCode != packet.ConnectAccess {
		log.Fatal("服务器拒绝连接，code:", connAckPacket.ConnectReturnCode)
		conn.Close()
	}
	fmt.Println("接收返回报文 connectAck, return code:", connAckPacket.ConnectReturnCode)

	b.IdentityPool = make(map[int]bool, )
	for i := 1; i <= math.MaxUint16; i++ {  // 非零的16位报文标识符
		b.IdentityPool[i] = true
	}
	return conn, nil
}
