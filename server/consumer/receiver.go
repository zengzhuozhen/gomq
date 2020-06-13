package consumer

import (
	"context"
	"fmt"
	"gomq/common"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"net"
	"time"
)

type Receiver struct {
	chanAssemble map[string][]common.MsgChan
	pool         *Pool
}

func NewConsumerReceiver(chanAssemble map[string][]common.MsgChan, pool *Pool) *Receiver {
	return &Receiver{chanAssemble: chanAssemble, pool: pool}
}

func (r *Receiver) HandleConn(netPacket common.Packet, conn net.Conn) {
	connUid := conn.RemoteAddr().String()
	r.pool.Add(connUid, netPacket.Topic, netPacket.Position)
	r.chanAssemble[connUid] = make([]common.MsgChan, 1000)
}

func (r *Receiver) HandleQuit(connUid string) {
	r.pool.State[connUid] = false
	for _, v := range r.chanAssemble[connUid] {
		close(v)
	}
	delete(r.chanAssemble, connUid)
}

//
//func (r *Receiver) Consume(ctx context.Context, conn net.Conn) {
//	connUid := conn.RemoteAddr().String()
//	for {
//		select {
//		case msg := <-r.chanAssemble[connUid]:
//			// 防止多个管道同时竞争所有消息的问题,采用客户端连接池进行逻辑隔离解决
//			netPacket := common.Packet{
//				Flag:    common.S,
//				Message: msg,
//				Topic:   r.pool.Topic[connUid],
//			}
//			if data, err := json.Marshal(netPacket); err == nil {
//				fmt.Printf("准备发送给客户端消费者: %s %+v", conn.RemoteAddr(), netPacket)
//				n, err := conn.Write(data)
//				fmt.Println(n)
//				if err != nil {
//					fmt.Println(err)
//				}
//			} else {
//				log.Fatalln(err.Error())
//			}
//		case <-ctx.Done():
//			fmt.Println("消费者连接已关闭，退出消费循环")
//			r.HandleQuit(connUid)
//			return
//		}
//	}
//}

func (r *Receiver) ConsumeAndResponse(ctx context.Context, conn net.Conn, packet *protocolPacket.SubscribePacket) {
	var TopicList []string
	var QoSList []byte
	for packet.RemainingLength > 0 {
		topic, _ := utils.DecodeString(conn)
		QoS, _ := utils.DecodeByte(conn)
		TopicList = append(TopicList, topic)
		QoSList = append(QoSList, QoS)
		packet.RemainingLength -= 2 + len(topic) + 1
	}
	subAckPacket := protocolPacket.NewSubAckPacket(packet.PacketIdentifier, QoSList)
	err := subAckPacket.Write(conn)
	if err != nil {
		fmt.Println("返回subAck失败")
	}
	connUid := conn.RemoteAddr().String()
	for {
		for k, v := range TopicList {
			go r.listenMsgChan(k, v, connUid, conn)
		}
		select {
		case <-ctx.Done():
			fmt.Println("消费者连接已关闭，退出消费循环")
			r.HandleQuit(connUid)
			return
		}
	}

}

func (r *Receiver) listenMsgChan(k int, topic, connUid string, conn net.Conn) {
	select {
	case msg := <-r.chanAssemble[connUid][k]:
		// 防止多个管道同时竞争所有消息的问题,采用客户端连接池进行逻辑隔离解决
		messagePacket := msg.Pack()
		if _, err := conn.Write(messagePacket); err != nil {
			fmt.Println(err)
		}
	}
}

func (r *Receiver) CloseConsumer(conn net.Conn, packet *protocolPacket.UnSubscribePacket) {

	connUid := conn.RemoteAddr().String()
	for packet.RemainingLength > 0 {
		topic, _ := utils.DecodeString(conn)
		// 这里根据consume 连接到server是提供的 topic 列表在pool中顺序排列的特点
		// 找出此次需要关闭的 topic 通道对应的 key ，需严格保证 pool 中所有数组顺序排列
		for k, top := range r.pool.Topic[connUid] {
			if top == topic {
				close(r.chanAssemble[connUid][k])
			}
		}
		packet.RemainingLength -= len(topic) + 2
	}

	// 返回取消订阅确认
	unSubAck := protocolPacket.NewUnSubAckPacket(packet.PacketIdentifier)
	unSubAck.Write(conn)
}

func (r *Receiver) Pong(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	pingRespPack := protocolPacket.NewPingRespPacket()
	pingRespPack.Write(conn)
}
