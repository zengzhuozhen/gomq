package service

import (
	"context"
	"fmt"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"gomq/server/consumer"
	"gomq/server/producer"
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

func (l *Listener) Start(producer *producer.Receiver, consumer *consumer.Receiver) {
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

func (l *Listener) holdConn(conn net.Conn, producerRc *producer.Receiver, consumerRc *consumer.Receiver) {
	// 如果没有可读数据，也就是读 buffer 为空，则阻塞
	ctx, cancel := context.WithCancel(context.Background())
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	for {
		packet, err := ReadPacket(conn)
		if err != nil {
			fmt.Println("客户端超时未响应或退出，关闭连接")
			_ = conn.Close()
			cancel()
			return
		}


		switch packet.(type) {
		case *protocolPacket.PublishPacket:
			producerRc.ProduceAndResponse(conn,packet.(*protocolPacket.PublishPacket))
		case *protocolPacket.SubscribePacket:
			consumerRc.ConsumeAndResponse(ctx,conn,packet.(*protocolPacket.SubscribePacket))
		case *protocolPacket.UnSubscribePacket:
			consumerRc.CloseConsumer(conn, packet.(*protocolPacket.UnSubscribePacket))
		case *protocolPacket.PingReqPacket:
			consumerRc.Pong(conn)
		case *protocolPacket.DisConnectPacket:
			cancel()
			conn.Close()
		}

		//data = data[:n]
		//netPacket := common.Packet{}
		//_ = json.Unmarshal(data, &netPacket)
		//switch netPacket.Flag {
		//case common.C: // 消费者连接
		//	consumerRc.HandleConn(netPacket, conn)
		//	go consumerRc.Consume(ctx, conn)
		//case common.CH: // 消费者心跳
		//	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
		//	continue
		//case common.P: // 生产者连接
		//	_ = conn.Close()
		//	return
		//}
	}
}

func getPackLen(bytes []byte, i int) (fieldLength, remainingLength int) {
	if bytes[0] > 128 {
		return getPackLen(bytes[1:], i+1)
	}
	return i, utils.BytesToInt(bytes[:i])
}

func handleConnectProtocol(conn net.Conn) bool {
	var connPacket protocolPacket.ConnectPacket
	var fh protocolPacket.FixedHeader
	initTime := time.Now()
	if err := fh.Read(conn);err != nil{
		fmt.Errorf("读取包头失败",err)
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
	if err := fh.Read(r);err != nil{
		fmt.Errorf("读取包头失败",err)
	}

	var packet protocolPacket.ControlPacket

	switch protocolPacket.DecodePacketType(fh.TypeAndReserved) {
	case byte(protocol.PUBLISH):
		//todo
		packet = &protocolPacket.PublishPacket{}
		packet.Read(r,fh)
	case byte(protocol.SUBSCRIBE):
		packet = &protocolPacket.SubscribePacket{}
		packet.Read(r,fh)
		// todo
	case byte(protocol.UNSUBSCRIBE):
		packet = &protocolPacket.UnSubscribePacket{}
		packet.Read(r,fh)
	case byte(protocol.PINGREQ):
		packet = &protocolPacket.PingReqPacket{}
		packet.Read(r,fh)
	}
	return packet,nil
}
