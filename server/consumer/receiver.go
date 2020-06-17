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
	ChanAssemble map[string][]common.MsgChan
	pool         *Pool
}

func NewConsumerReceiver(chanAssemble map[string][]common.MsgChan, pool *Pool) *Receiver {
	return &Receiver{ChanAssemble: chanAssemble, pool: pool}
}

func (r *Receiver) HandleQuit(connUid string) {
	r.pool.State[connUid] = false
	for _, v := range r.ChanAssemble[connUid] {
		if v != nil {
			close(v)
		}
	}
	delete(r.ChanAssemble, connUid)
}

func (r *Receiver) ConsumeAndResponse(ctx context.Context, conn net.Conn, packet *protocolPacket.SubscribePacket) {
	var TopicList []string
	var QoSList []byte
	packet.PacketIdentifier, _ = utils.DecodeUint16(conn)
	packet.FixedHeader.RemainingLength -= 2
	for packet.FixedHeader.RemainingLength > 0 {
		topic, err := utils.DecodeString(conn)
		if err != nil {
			fmt.Println(err)
		}
		QoS, err := utils.DecodeByte(conn)
		if err != nil {
			fmt.Println(err)
		}
		TopicList = append(TopicList, topic)
		QoSList = append(QoSList, QoS)
		packet.RemainingLength -= 2 + len(topic) + 1
	}

	connUid := conn.RemoteAddr().String()
	r.pool.Add(connUid, TopicList)
	subAckPacket := protocolPacket.NewSubAckPacket(packet.PacketIdentifier, QoSList)
	err := subAckPacket.Write(conn)
	if err != nil {
		fmt.Println("返回subAck失败", err)
	}
	for k, v := range TopicList {
		tempChan := make(common.MsgChan)
		r.ChanAssemble[connUid] = append(r.ChanAssemble[connUid], tempChan)
		fmt.Println(r.ChanAssemble)
		go r.listenMsgChan(k, v, connUid, conn)
	}
	go func(connUid string) {
		select {
		case <-ctx.Done():
			fmt.Println("消费者连接已关闭，退出消费循环")
			r.HandleQuit(connUid)
			return
		}
	}(connUid)

}

func (r *Receiver) listenMsgChan(k int, topic, connUid string, conn net.Conn) {


	for {
		select {
		case msg := <-r.ChanAssemble[connUid][k]:
			// 防止多个管道同时竞争所有消息的问题,采用客户端连接池进行逻辑隔离解决
			messagePacket := msg.Pack()
			fmt.Println("准备推送消息", messagePacket)
			if _, err := conn.Write(messagePacket); err != nil {
				fmt.Println(err)
			}
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
				close(r.ChanAssemble[connUid][k])
			}
		}
		packet.RemainingLength -= len(topic) + 2
	}

	// 返回取消订阅确认
	unSubAck := protocolPacket.NewUnSubAckPacket(packet.PacketIdentifier)
	unSubAck.Write(conn)
}

func (r *Receiver) Pong(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	pingRespPack := protocolPacket.NewPingRespPacket()
	err := pingRespPack.Write(conn)
	if err != nil{
		fmt.Println(err)
	}
}
