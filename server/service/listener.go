package service

import (
	"context"
	"fmt"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"io"
	"log"
	"net"
	"reflect"
	"time"
)

type Listener struct {
	protocol string
	address  string
}

func NewListener(protocol, address string) *Listener {
	return &Listener{
		protocol: protocol,
		address:  address,
	}
}

func (l *Listener) Start(producer *ProducerReceiver, consumer *ConsumerReceiver) {
	listen, err := net.Listen(l.protocol, l.address)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for {
		conn, err := listen.Accept()
		fmt.Println("客户端：" + conn.RemoteAddr().String() + "连接")
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		if handleConnectProtocol(conn) == true {
			go l.holdConn(conn, producer, consumer)
		}
	}
}

func (l *Listener) holdConn(conn net.Conn, producerRc *ProducerReceiver, consumerRc *ConsumerReceiver) {
	// 如果没有可读数据，也就是读 buffer 为空，则阻塞
	ctx, cancel := context.WithCancel(context.Background())
	_ = conn.SetDeadline(time.Now().Add(1000 * time.Second))
	for {
		packet, err := ReadPacket(conn)
		if err != nil {
			fmt.Println("客户端超时未响应或退出，关闭连接" + err.Error())
			_ = conn.Close()
			cancel()
			return
		}

		switch packet.(type) {
		case *protocolPacket.PublishPacket:
			producerRc.ProduceAndResponse(conn, packet.(*protocolPacket.PublishPacket))
		case *protocolPacket.SubscribePacket:
			consumerRc.ConsumeAndResponse(ctx, conn, packet.(*protocolPacket.SubscribePacket))
		case *protocolPacket.UnSubscribePacket:
			consumerRc.CloseConsumer(conn, packet.(*protocolPacket.UnSubscribePacket))
		case *protocolPacket.PingReqPacket:
			consumerRc.Pong(conn)
		case *protocolPacket.DisConnectPacket:
			cancel()
			conn.Close()
		}
	}
}

func handleConnectProtocol(conn net.Conn) bool {
	var connPacket protocolPacket.ConnectPacket
	var fh protocolPacket.FixedHeader
	initTime := time.Now()
	if err := fh.Read(conn); err != nil {
		fmt.Errorf("读取包头失败", err)
	}
	err := connPacket.Read(conn, fh)
	if err != nil || time.Now().Sub(initTime) > 3*time.Second {
		fmt.Println("客户端断开连接或未再规定时间内发送消息")
		conn.Close()
		return false
	}

	//----判断是否满足协议规范-----

	if !connPacket.IsReserved() {
		fmt.Println("客户端连接标识错误")
		conn.Close()
		return false
	}
	if reflect.DeepEqual(connPacket.TypeAndReserved, utils.EncodeString("MQTT")) {
		fmt.Println("客户端使用的协议错误")
		conn.Close()
		return false
	}

	//----以下为满足报文规范的情况---

	if !connPacket.IsLegalIdentifier() {
		responseConnectAck(conn, protocolPacket.UnSupportClientIdentity)
		fmt.Println("客户端唯一标识错误")
		conn.Close()
		return false
	}

	if ! connPacket.IsAuthorizedClient() {
		responseConnectAck(conn, protocolPacket.UnAuthorization)
		fmt.Println("客户端未授权")
		conn.Close()
		return false
	}

	if !connPacket.IsCorrectSecret() {
		responseConnectAck(conn, protocolPacket.UserAndPassError)
		fmt.Println("客户端user和password错误")
		conn.Close()
		return false
	}

	// ------以下为处理 payLoad 里的数据
	// todo

	// 返回ack
	responseConnectAck(conn, protocolPacket.ConnectAccess)
	return true
}

func responseConnectAck(conn net.Conn, code byte) {
	connectAckPacket := protocolPacket.NewConnectAckPack(code)
	connectAckPacket.Write(conn)
}

// 读取数据包
func ReadPacket(r io.Reader) (protocolPacket.ControlPacket, error) {
	var fh protocolPacket.FixedHeader
	if err := fh.Read(r); err != nil {
		fmt.Errorf("读取包头失败", err)
		return nil,err
	}

	var packet protocolPacket.ControlPacket

	switch protocolPacket.DecodePacketType(fh.TypeAndReserved) {
	case byte(protocol.PUBLISH):
		packet = &protocolPacket.PublishPacket{}
		packet.Read(r, fh)
	case byte(protocol.SUBSCRIBE):
		packet = &protocolPacket.SubscribePacket{}
		packet.ReadHeadOnly(r, fh)
	case byte(protocol.UNSUBSCRIBE):
		packet = &protocolPacket.UnSubscribePacket{}
		packet.ReadHeadOnly(r, fh)
	case byte(protocol.PINGREQ):
		packet = &protocolPacket.PingReqPacket{}
		packet.Read(r, fh)
	}
	return packet, nil
}
