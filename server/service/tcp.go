package service

import (
	"context"
	"fmt"
	"gomq/protocol"
	"gomq/protocol/handler"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"io"
	"log"
	"net"
	"time"
)

type TCP struct {
	protocol string
	address  string
	*ProducerReceiver
	*ConsumerReceiver
	*MemberReceiver
}

func NewTCP(address string, PR *ProducerReceiver, CR *ConsumerReceiver, MR *MemberReceiver) *TCP {
	return &TCP{
		protocol:         "tcp",
		address:          address,
		ProducerReceiver: PR,
		ConsumerReceiver: CR,
		MemberReceiver:   MR,
	}
}

func (tcp *TCP) Start() {
	go tcp.startConnLoop()
	listen, err := net.Listen(tcp.protocol, tcp.address)
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
			go tcp.holdConn(conn)
		}
	}
}

func (tcp *TCP) startConnLoop() error {
	fmt.Println("开启监听连接循环")
	for {
		activeConn := tcp.ConsumerReceiver.Pool.ForeachActiveConn()
		if len(activeConn) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, uid := range activeConn {
			topicList := tcp.ConsumerReceiver.Pool.Topic[uid]
			for k, topic := range topicList {
				tcp.popRetainQueue(uid, topic, k)
				tcp.popQueue(uid, topic, k)
			}
		}
	}
}

func (tcp *TCP) popRetainQueue(uid, topic string, k int) {
	position := tcp.ConsumerReceiver.Pool.Position[uid][k]
	if position == 0 { // 没有偏移,即是新来的, 直接读retaineQueue，然后更新到最新偏移
		fmt.Println(tcp.ProducerReceiver.Queue.Local[topic])
		for _, msg := range tcp.ProducerReceiver.RetainQueue.ReadAll(topic) {
			tcp.ConsumerReceiver.ChanAssemble[uid][k] <- msg
		}
		maxPosition := len(tcp.ProducerReceiver.Queue.Local[topic])
		// 更新到最新的偏移
		tcp.ConsumerReceiver.Pool.UpdatePositionTo(uid, topic, maxPosition)
	}
}

func (tcp *TCP) popQueue(uid, topic string, k int) {
	position := tcp.ConsumerReceiver.Pool.Position[uid][k]
	// 正常处理
	if msg, err := tcp.ProducerReceiver.Queue.Pop(topic, position); err == nil {
		tcp.ConsumerReceiver.Pool.UpdatePosition(uid, topic)
		tcp.ConsumerReceiver.ChanAssemble[uid][k] <- msg
	} else {
		time.Sleep(100 * time.Millisecond)
	}
}

func (tcp *TCP) holdConn(conn net.Conn) {
	// 如果没有可读数据，也就是读 buffer 为空，则阻塞
	ctx, cancel := context.WithCancel(context.Background())
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
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
			tcp.ProducerReceiver.ProduceAndResponse(conn, packet.(*protocolPacket.PublishPacket))
		case *protocolPacket.PubRelPacket:
			tcp.ProducerReceiver.AcceptRelAndRespComp(conn, packet.(*protocolPacket.PubRelPacket))
		case *protocolPacket.SubscribePacket:
			tcp.ConsumerReceiver.ConsumeAndResponse(ctx, conn, packet.(*protocolPacket.SubscribePacket))
		case *protocolPacket.UnSubscribePacket:
			tcp.ConsumerReceiver.CloseConsumer(conn, packet.(*protocolPacket.UnSubscribePacket))
		case *protocolPacket.PingReqPacket:
			tcp.ConsumerReceiver.Pong(conn)
		case *protocolPacket.SyncReqPacket:
			tcp.MemberReceiver.RegisterSyncAndResponse(conn, packet.(*protocolPacket.SyncReqPacket))
		case *protocolPacket.SyncOffsetPacket:
			tcp.MemberReceiver.UpdateSyncOffset(conn, packet.(*protocolPacket.SyncOffsetPacket))
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

	connectFlags, payLoad := connPacket.ProvisionConnectFlagsAndPayLoad()
	connectPacketHandler := handler.NewConnectPacketHandle(&connPacket, connectFlags, payLoad)
	err = connectPacketHandler.HandleAll()
	if err != nil {
		conn.Close()
		// 判断错误类型，适时发responseAck
		errorType := connectPacketHandler.ErrorTypeToAck()
		if errorType != 0 {
			responseConnectAck(conn, errorType)
		}
		return false
	}

	// 返回ack
	responseConnectAck(conn, protocol.ConnectAccess)
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
		return nil, err
	}

	var packet protocolPacket.ControlPacket
	var err error

	switch fh.TypeAndReserved {
	// 自定义的协议包，无须解包
	case protocol.SYNCREQ:
		packet = &protocolPacket.SyncReqPacket{}
		err = packet.Read(r, fh)
	case protocol.SYNCACK:
		packet = &protocolPacket.SyncAckPacket{}
		err = packet.Read(r, fh)
	case protocol.SYNCOFFSET:
		packet = &protocolPacket.SyncOffsetPacket{}
		err = packet.Read(r, fh)
	}
	// MQTT协议包，需要做一层解包
	switch utils.DecodePacketType(fh.TypeAndReserved) {
	case byte(protocol.PUBLISH):
		packet = &protocolPacket.PublishPacket{}
		err = packet.Read(r, fh)
	case byte(protocol.PUBREL):
		packet = &protocolPacket.PubRelPacket{}
		err = packet.Read(r,fh)
	case byte(protocol.SUBSCRIBE):
		packet = &protocolPacket.SubscribePacket{}
		err = packet.ReadHeadOnly(r, fh) // 只读头部，剩下的具体业务里面处理
	case byte(protocol.UNSUBSCRIBE):
		packet = &protocolPacket.UnSubscribePacket{}
		err = packet.ReadHeadOnly(r, fh) // 只读头部，剩下的具体业务里面处理
	case byte(protocol.PINGREQ):
		packet = &protocolPacket.PingReqPacket{}
		err = packet.Read(r, fh)
	case byte(protocol.DISCONNECT):
		packet = &protocolPacket.DisConnectPacket{}
		err = packet.Read(r, fh)
	}
	return packet, err
}
