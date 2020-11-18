package service

import (
	"errors"
	"fmt"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"net"
	"sync"
)

type ProducerReceiver struct {
	Queue *Queue
}

func NewProducerReceiver() *ProducerReceiver {
	return &ProducerReceiver{
		&Queue{
			local:          make(map[string][]common.MessageUnit, 1024),
			offset:         make(map[string]int64),
			mu:             new(sync.RWMutex),
			PersistentChan: make(chan common.MessageUnit),
		},
	}
}

func (p *ProducerReceiver) ProduceAndResponse(conn net.Conn, publishPacket *protocolPacket.PublishPacket) {
	bit4 := publishPacket.TypeAndReserved - 16 - 32 // 去除 MQTT协议类型
	var needHandleRetain bool
	if bit4 >= 8 { //重发标志 DUP, 0 表示第一次发这个消息
		bit4 -= 8
	}
	if bit4%2 == 1 { //保留标志 RETAIN ,为1 需要保存消息和服务等级
		needHandleRetain = true
		bit4--
	}

	switch bit4 { // 服务质量等级 QoS，左移1位越过retain
	case protocol.AtMostOnce:
		// nothing to do
	case protocol.AtLeastOnce << 1:
		defer responsePubAck(conn, publishPacket.PacketIdentifier)
	case protocol.ExactOnce << 1:
		defer responsePubRec(conn, publishPacket.PacketIdentifier)
	case protocol.None << 1:
		// nothing to do
	}

	if needHandleRetain {
		protocolPacket.HandleRetain()
	}
	message := new(common.Message)
	message = message.UnPack(publishPacket.Payload)
	messageUnit := common.NewMessageUnit(publishPacket.TopicName, *message)
	fmt.Printf("主题 %s 生产了: %s ", publishPacket.TopicName, message.MsgKey)
	p.Queue.Push(messageUnit)

	fmt.Println("记录入队数据", message.MsgKey)

}

func responsePubAck(conn net.Conn, identify uint16) {
	fmt.Println("发送puback")
	pubAckPacket := protocolPacket.NewPubAckPacket(identify)
	pubAckPacket.Write(conn)
}

func responsePubRec(conn net.Conn, identify uint16) {
	fmt.Println("准备返回pubRec")
	// PUBREC – 发布收到（QoS 2，第一步)
	pubRecPacket := protocolPacket.NewPubRecPacket(identify)
	pubRecPacket.Write(conn)

	// 等待 rel
	var fh protocolPacket.FixedHeader
	var pubRelPacket protocolPacket.PubRelPacket
	if err := fh.Read(conn); err != nil {
		fmt.Println("接收pubRel包头内容错误", err)
	}
	pubRelPacket.Read(conn, fh)

	// PUBCOMP – 发布完成（QoS 2，第三步)
	fmt.Println("准备返回pubComp")
	pubCompPacket := protocolPacket.NewPubCompPacket(pubRelPacket.PacketIdentifier)
	pubCompPacket.Write(conn)
	// todo identify Pool
	return
}

type Queue struct {
	local          map[string][]common.MessageUnit // local[Topic][Message_1,Message_2]
	offset         map[string]int64                // offset[Topic]1024
	mu             *sync.RWMutex
	PersistentChan chan common.MessageUnit
}

func (q *Queue) Push(messageUnit common.MessageUnit) {
	q.mu.RLock()
	if len(q.local[messageUnit.Topic]) >= 1024 { // 超过 1024个元素 清空一下
		q.offset[messageUnit.Topic] += int64(len(q.local[messageUnit.Topic]))
		q.local[messageUnit.Topic] = make([]common.MessageUnit, 0)
	}
	q.local[messageUnit.Topic] = append(q.local[messageUnit.Topic], messageUnit)
	q.mu.RUnlock()
	q.PersistentChan <- messageUnit

}

func (q *Queue) Pop(topic string, position int64) (message common.MessageUnit, err error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	position -= q.offset[topic]
	// todo
	// 新加进来的消费者会从第一个开始消费，此时 position < offset , 需要实现类似缺页中断的功能，
	// 并且把之后所有的message都放在一个临时buffer里，(局部性原理)
	if len(q.local[topic]) == 0 {
		return common.MessageUnit{}, errors.New("目前还没有消息")
	} else if int64(len(q.local[topic])) <= position {
		return common.MessageUnit{}, errors.New("已消费到最新消息")
	} else {
		return q.local[topic][position], nil
	}
}
