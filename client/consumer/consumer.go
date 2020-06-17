package consumer

import (
	"fmt"
	"gomq/client"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"net"
	"time"
)

type Consumer struct {
	bc *client.BasicClient
}

func NewConsumer(protocol, host string, port, timeout int) *Consumer {
	return &Consumer{bc: &client.BasicClient{
		Protocol: protocol,
		Host:     host,
		Port:     port,
		Timeout:  timeout,
	}}
}

func (c *Consumer) Subscribe(topic []string) {
	conn, err := c.bc.Connect()
	fmt.Println(conn.LocalAddr().String())
	if err != nil {
		panic("连接服务端失败")
	}
	// 连接服务端
	subscribePacket := protocolPacket.NewSubscribePacket(1, topic, 0)
	err = subscribePacket.Write(conn)
	if err != nil {
		fmt.Println("客户端订阅主题失败:发送subscribe")
	}
	fmt.Println("发送subscribe")

	go c.Heart(conn, 3 * time.Second)
	for {
		c.ReadPacket(conn)
	}
}

func (c *Consumer) Heart(conn net.Conn, duration time.Duration){
	func() {
		tickTimer := time.NewTicker(duration)
		for {
			select {
			case <-tickTimer.C:
				fmt.Println("发送心跳包")
				pingReqPack := protocolPacket.NewPingReqPacket()
				pingReqPack.Write(conn)
			}
		}
	}()
}

func (c *Consumer) ReadPacket(r net.Conn) {
	// 读取数据包
	var fh protocolPacket.FixedHeader
	if err := fh.Read(r); err != nil {
		fmt.Errorf("读取包头失败%+v", err)
		return
	}

	switch protocolPacket.DecodePacketType(fh.TypeAndReserved) {
	case byte(protocol.SUBACK):
		var subAckPacket protocolPacket.SubAckPacket
		err := subAckPacket.Read(r, fh)
		if err != nil {
			fmt.Println("客户端订阅主题失败:等待subAck")
			r.Close()
		}
		fmt.Println("收到服务端确认订阅消息")
	case byte(protocol.UNSUBACK):
		var fh protocolPacket.FixedHeader
		var unSubAckPacket protocolPacket.UnSubAckPacket
		_ = fh.Read(r)
		unSubAckPacket.Read(r, fh)
		fmt.Println("收到服务端确认取消订阅")
	case byte(protocol.PINGRESP):
		fmt.Println("收到服务端心跳回应")
	default:
		// 普通消息
		var messByte [1000]byte
		n, _ :=r.Read(messByte[:])

		head := fh.Pack()
		data := append(head.Bytes(),messByte[:n]...)
		message := new(common.Message)
		message = message.UnPack(data)
		fmt.Printf("收到服务端消息%+v",message)
	}
}

func (c *Consumer) UnSubscribe(conn net.Conn, topic []string, identifier uint16) {
	unSubscribePack := protocolPacket.NewUnSubscribePacket(identifier, topic)
	_ = unSubscribePack.Write(conn)
}

func (c *Consumer) DisConnect(conn net.Conn) {
	disConnectPack := protocolPacket.NewDisConnectPacketPacket()
	disConnectPack.Write(conn)
	conn.Close()
}
