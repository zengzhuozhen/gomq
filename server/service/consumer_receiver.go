package service

import (
	"context"
	"fmt"
	"gomq/common"
	protocolPacket "gomq/protocol/packet"
	"gomq/protocol/utils"
	"net"
	"strings"
	"sync"
	"time"
)

type QoSForTopic map[string]byte

type ConsumerReceiver struct {
	ChanAssemble map[string][]common.MsgUnitChan
	QoSGuarantee map[string]QoSForTopic
	Pool         *Pool
}

func NewConsumerReceiver(chanAssemble map[string][]common.MsgUnitChan) *ConsumerReceiver {
	return &ConsumerReceiver{
		ChanAssemble: chanAssemble,
		QoSGuarantee: make(map[string]QoSForTopic),
		Pool: &Pool{
			State:    new(sync.Map),
			Position: make(map[string][]int64, 1024),
			Topic:    make(map[string][]string, 1024),
			IsOldOne: make(map[string]bool),
			mu:       sync.Mutex{}},

	}
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

	connUid := conn.RemoteAddr().String()
	qosForTopic := make(map[string]byte)

	for packet.FixedHeader.RemainingLength > 0 {
		topic, err := utils.DecodeString(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}
		QoS, err := utils.DecodeByte(conn)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if strings.HasSuffix(topic, "*") {
			fmt.Println("目前服务端不支持通配符")
			continue
		}

		TopicList = append(TopicList, topic)
		QoSList = append(QoSList, QoS)
		qosForTopic[topic] = QoS
		packet.RemainingLength -= 2 + len(topic) + 1 // LSB + MSB + topic + qos
	}

	r.QoSGuarantee[connUid] = qosForTopic
	r.Pool.Add(connUid, TopicList)
	subAckPacket := protocolPacket.NewSubAckPacket(packet.PacketIdentifier, QoSList)
	err := subAckPacket.Write(conn)
	if err != nil {
		fmt.Println("返回subAck失败", err)
	}

	for topicIndex := range TopicList {
		r.ChanAssemble[connUid] = append(r.ChanAssemble[connUid], make(common.MsgUnitChan))
		go r.listenMsgChan(ctx, topicIndex, connUid, conn)
	}

}

func (r *ConsumerReceiver) listenMsgChan(ctx context.Context, topicIndex int, connUid string, conn net.Conn) {
	for {
		select {
		case msg := <-r.ChanAssemble[connUid][topicIndex]:
			// 防止多个管道同时竞争所有消息的问题,采用客户端连接池进行逻辑隔离解决
			messagePacket := msg.Pack()
			messagePacket = append(messagePacket,[]byte{'\n'}...)
			fmt.Printf("准备推送消息:{Topic:'%s'} {Body:'%s'}", msg.Topic, msg.Data.Body)
			if _, err := conn.Write(messagePacket); err != nil {
				fmt.Println(err)
			}
			// todo choose send message	strategy by qos guarantee


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
		for topicIndex, top := range r.Pool.Topic[connUid] {
			if top == topic {
				fmt.Println("关闭", connUid, "的主题", topic)
				close(r.ChanAssemble[connUid][topicIndex])
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
	State *sync.Map
	// position 和 topic 总是相对应的
	// Position[ConnId]{Topic_A_position,Topic_B_position,Topic_C_position}
	//							｜				｜				｜
	// Topic[ConnId]   {	Topic_A,		Topic_B,		Topic_C		}
	Position map[string][]int64
	Topic    map[string][]string
	IsOldOne map[string]bool
	mu       sync.Mutex
}

func (p *Pool) ForeachActiveConn() []string {
	connUids := make([]string, 0)
	p.State.Range(func(connUid, isActive interface{}) bool {
		if isActive.(bool) == true && len(p.Topic[connUid.(string)]) != 0 {
			connUids = append(connUids, connUid.(string))
		}
		return true
	})
	return connUids
}

func (p *Pool) Add(connUid string, topics []string) {
	p.State.Store(connUid, true)
	p.Position[connUid] = make([]int64, len(topics))
	p.Topic[connUid] = topics
}

func (p *Pool) UpdatePosition(uid, topic string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range p.Topic[uid] {
		if topic == v {
			p.Position[uid][k]++
			break
		}
	}
}

func (p *Pool) UpdatePositionTo(uid, topic string, to int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range p.Topic[uid] {
		if topic == v {
			p.Position[uid][k] = int64(to)
			break
		}
	}
}
