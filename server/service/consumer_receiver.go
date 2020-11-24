package service

import (
	"context"
	"fmt"
	"gomq/common"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"net"
	"sync"
	"time"
)

type ConsumerReceiver struct {
	ChanAssemble map[string][]common.MsgUnitChan
	Pool         *Pool
}

func NewConsumerReceiver(chanAssemble map[string][]common.MsgUnitChan) *ConsumerReceiver {
	return &ConsumerReceiver{ChanAssemble: chanAssemble, Pool: &Pool{
		ConnUids: []string{},
		State:    new(sync.Map),
		Position: make(map[string][]int64, 1024),
		Topic:    make(map[string][]string, 1024),
		mu:       sync.Mutex{},
	}}
}

func (r *ConsumerReceiver) HandleQuit(connUid string) {
	r.Pool.State.Store(connUid, false)
	for _, v := range r.ChanAssemble[connUid] {
		_, isClose := <-v
		if v != nil && !isClose {
			close(v)
		}
	}
	delete(r.ChanAssemble, connUid)
}

func (r *ConsumerReceiver) ConsumeAndResponse(ctx context.Context, conn net.Conn, packet *protocolPacket.SubscribePacket) {
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
	r.Pool.Add(connUid, TopicList)
	subAckPacket := protocolPacket.NewSubAckPacket(packet.PacketIdentifier, QoSList)
	err := subAckPacket.Write(conn)
	if err != nil {
		fmt.Println("返回subAck失败", err)
	}
	for k := range TopicList {
		r.ChanAssemble[connUid] = append(r.ChanAssemble[connUid], make(common.MsgUnitChan))
		go r.listenMsgChan(ctx, k, connUid, conn)
	}
}

func (r *ConsumerReceiver) listenMsgChan(ctx context.Context, k int, connUid string, conn net.Conn) {
	for {
		select {
		case msg := <-r.ChanAssemble[connUid][k]:
			// 防止多个管道同时竞争所有消息的问题,采用客户端连接池进行逻辑隔离解决
			messagePacket := msg.Pack()
			fmt.Printf("准备推送消息:{Topic:'%s'} {Body:'%s'}", msg.Topic, msg.Data.Body)
			if _, err := conn.Write(messagePacket); err != nil {
				fmt.Println(err)
			}
		case <-ctx.Done():
			fmt.Printf("客户端{socket:'%s'}连接关闭，退出消费", connUid)
			r.HandleQuit(connUid)
			return
		}
	}
}

func (r *ConsumerReceiver) CloseConsumer(conn net.Conn, packet *protocolPacket.UnSubscribePacket) {
	connUid := conn.RemoteAddr().String()
	for packet.RemainingLength > 0 {
		topic, _ := utils.DecodeString(conn)
		// 这里根据consume 连接到server是提供的 topic 列表在pool中顺序排列的特点
		// 找出此次需要关闭的 topic 通道对应的 key ，需严格保证 Pool 中所有数组顺序排列
		for k, top := range r.Pool.Topic[connUid] {
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

func (r *ConsumerReceiver) Pong(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	pingRespPack := protocolPacket.NewPingRespPacket()
	err := pingRespPack.Write(conn)
	if err != nil {
		fmt.Println(err)
	}
}

type Pool struct {
	ConnUids []string
	State    *sync.Map
	// position 和 topic 总是相对应的
	Position map[string][]int64  // Position[ConnId]{Topic_A_position,Topic_B_position,Topic_C_position}
	Topic    map[string][]string // Topic[ConnId]{Topic_A,Topic_B,Topic_C}
	mu       sync.Mutex
}

func (p *Pool) ForeachActiveConn() []string {
	connUids := make([]string, 0)
	for _, i := range p.ConnUids {
		isActive, ok := p.State.Load(i)
		if !ok {
			panic("active not exist in state Pool")
		}
		if isActive == true && len(p.Topic[i]) != 0 {
			connUids = append(connUids, i)
		}
	}
	return connUids
}

func (p *Pool) Add(connUid string, topics []string) {
	p.ConnUids = append(p.ConnUids, connUid)
	p.State.Store(connUid, true)
	p.Position[connUid] = make([]int64, len(topics))
	p.Topic[connUid] = topics
}

func (p *Pool) UpdatePosition(uid, topic string) {
	p.mu.Lock()
	for k, v := range p.Topic[uid] {
		if topic == v {
			p.Position[uid][k]++
			break
		}
	}
	p.mu.Unlock()
}