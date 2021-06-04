package service

import (
	"context"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/protocol"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"github.com/zengzhuozhen/gomq/protocol/visit"
	"io"
	"net"
	"sync"
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

// Start 开启TCP服务
func (tcp *TCP) Start() {
	go tcp.startConnLoop()
	go tcp.startCleanMessage()
	listen, err := net.Listen(tcp.protocol, tcp.address)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	for {
		conn, err := listen.Accept()
		log.Debugf("客户端：" + conn.RemoteAddr().String() + "连接")
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		if tcp.handleConnectProtocol(conn) == true {
			go tcp.holdConn(conn)
		}
	}
}

// startConnLoop 开始监听连接客户端，每次循环都会尝试推送消息给活跃的客户端
func (tcp *TCP) startConnLoop() {
	log.Infof("开启监听连接循环")
	for {
		activeConn := tcp.ConsumerReceiver.Pool.ForeachActiveConn()
		if len(activeConn) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, uid := range activeConn {
			topicList := tcp.ConsumerReceiver.Pool.Connections[uid].ConsumerConnAbs.Topic
			var wg sync.WaitGroup
			for topicIndex, topic := range topicList {
				wg.Add(1)
				go func(topicIndex int, topic string) {
					tcp.popRetainQueue(uid, topic, topicIndex)
					tcp.popQueue(uid, topic, topicIndex)
					wg.Done()
				}(topicIndex, topic)
			}
			wg.Wait()
		}
	}
}

// startCleanMessage 开启清理无用消息的goroutine
func (tcp *TCP) startCleanMessage() {
	log.Infof("开启周期消息清理机制")
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Infof("开始清理无用Message")
			topicMinOffset := make(map[string]int64)
			// 遍历每个topic的position，找出全局最小的pos
			for _, connection := range tcp.Pool.Connections {
				if connection.ConsumerConnAbs != nil {
					for top, pos := range connection.ConsumerConnAbs.TopPosMap {
						if topicMinOffset[top] == 0 || topicMinOffset[top] > pos {
							topicMinOffset[top] = pos
						}
					}
				}
			}
			log.Infof("目前消息内存占用情况", tcp.ProducerReceiver.Queue.GetLen("A"))
			queue := tcp.ProducerReceiver.Queue
			queue.Mutex.Lock()
			tcp.Pool.mu.Lock()
			for top, pos := range topicMinOffset {
				queue.Local[top] = queue.Local[top][pos:] // 清除Queue中不再使用的Message,这里只减少了长度，GC的时候没有回收对应空间，等下一次push的时候才会更新内存
				for _, connection := range tcp.Pool.Connections {
					if connection.ConsumerConnAbs != nil{
						connection.ConsumerConnAbs.TopPosMap[top] -= pos // 更新对应的Position
					}
				}
			}
			tcp.Pool.mu.Unlock()
			queue.Mutex.Unlock()
			log.Infof("清理后消息内存占用情况", tcp.ProducerReceiver.Queue.GetLen("A"))
		}
	}
}

func (tcp *TCP) popRetainQueue(uid, topic string, topicIndex int) {
	if !tcp.ConsumerReceiver.Pool.Connections[uid].IsOldOne {
		// todo 优化，每次都要去查一遍保留队列是否为空，并且是IO操作，考虑用cache标识是否retain里面有值
		if tcp.ProducerReceiver.RetainQueue.Cap(topic) > 0 { // 读一下retainQueue的保留内容
			msgList := tcp.ProducerReceiver.RetainQueue.ReadAll(topic)
			for _, msg := range msgList {
				tcp.ConsumerReceiver.ChanAssemble[uid][topicIndex] <- msg
			}
		}
		maxPosition := len(tcp.ProducerReceiver.Queue.Local[topic])
		// 更新到最新的偏移
		tcp.ConsumerReceiver.Pool.UpdatePositionTo(uid, topic, maxPosition)
		tcp.ConsumerReceiver.Pool.Connections[uid].IsOldOne = true
	}
}

func (tcp *TCP) popQueue(uid, topic string, topicIndex int) {
	position := tcp.ConsumerReceiver.Pool.Connections[uid].ConsumerConnAbs.TopPosMap[topic]
	// 正常处理
	if msg, err := tcp.ProducerReceiver.Queue.Pop(topic, position); err == nil {
		tcp.ConsumerReceiver.Pool.UpdatePosition(uid, topic)
		tcp.ConsumerReceiver.ChanAssemble[uid][topicIndex] <- msg
	} else {
		time.Sleep(100 * time.Millisecond)
	}
}

func (tcp *TCP) holdConn(conn net.Conn) {
	// 如果没有可读数据，也就是读 buffer 为空，则阻塞
	ctx, cancel := context.WithCancel(context.Background())
	keepAlive := tcp.ConsumerReceiver.Pool.Connections[conn.RemoteAddr().String()].KeepAlive
	_ = conn.SetDeadline(time.Now().Add(time.Duration(keepAlive) * time.Second))
	for {
		packet, err := ReadPacket(conn)
		if err != nil {
			log.Debugf("客户端超时未响应或退出，关闭连接" + err.Error())
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
			tcp.ConsumerReceiver.UnSubscribeAndResponse(conn, packet.(*protocolPacket.UnSubscribePacket))
		case *protocolPacket.PingReqPacket:
			tcp.ConsumerReceiver.Pong(conn)
		case *protocolPacket.SyncReqPacket:
			tcp.MemberReceiver.RegisterSyncAndResponse(conn, packet.(*protocolPacket.SyncReqPacket))
		case *protocolPacket.SyncOffsetPacket:
			tcp.MemberReceiver.UpdateSyncOffset(conn, packet.(*protocolPacket.SyncOffsetPacket))
		case *protocolPacket.DisConnectPacket:
			_ = conn.Close()
			delete(tcp.Pool.Connections,conn.RemoteAddr().String())
			cancel()
			return
		}
	}
}

func (tcp *TCP) handleConnectProtocol(conn net.Conn) bool {
	var connPacket protocolPacket.ConnectPacket
	var fh protocolPacket.FixedHeader
	initTime := time.Now()
	if err := fh.Read(conn); err != nil {
		log.Errorf("读取固定包头失败", err)
	}
	err := connPacket.Read(conn, fh)
	if err != nil || time.Now().Sub(initTime) > 3*time.Second {
		log.Errorf("客户端断开连接或未再规定时间内发送消息")
		conn.Close()
		return false
	}
	tcp.ConsumerReceiver.Pool.Connections[conn.RemoteAddr().String()] = &common.ConnectionAbstract{}
	if err = visit.NewConnectPacketVisitor(&visit.PacketVisitor{Packet: &connPacket}, tcp.ConsumerReceiver.Pool.Connections[conn.RemoteAddr().String()]).
		Visit(func(packet protocolPacket.ControlPacket) error { // 正常连接，返回连接成功ack
			responseConnectAck(conn, protocol.ConnectAccess)
			return nil
		}); err != nil {
		connectError, ok := err.(visit.ConnectError)
		if ok { // 发送连接失败错误码
			responseConnectAck(conn, byte(connectError.Code))
			return false
		}
	}
	return true
}

func responseConnectAck(conn net.Conn, code byte) {
	connectAckPacket := protocolPacket.NewConnectAckPack(code)
	connectAckPacket.Write(conn)
}

// ReadPacket 读取数据包
func ReadPacket(r io.Reader) (protocolPacket.ControlPacket, error) {
	var fh protocolPacket.FixedHeader
	if err := fh.Read(r); err != nil && err != io.EOF {
		log.Debugf("读取固定包头失败")
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
		err = packet.Read(r, fh)
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
