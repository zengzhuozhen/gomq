package producer

import (
	"fmt"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
	"net"
)

type Receiver struct {
	msgChan chan<- common.Message
	queue   *common.Queue
}

func NewProducerReceiver(msgChan chan common.Message, queue *common.Queue) *Receiver {
	return &Receiver{
		msgChan: msgChan,
		queue:   queue,
	}
}

func (p *Receiver) Produce(topic string, msg common.Message) {
	fmt.Printf("主题 %s 生产了: %s ", topic, msg.MsgKey)
	p.queue.Push(topic, msg)
	p.msgChan <- msg
}

func (p *Receiver) ProduceAndResponse(conn net.Conn ,publishPacket *protocolPacket.PublishPacket) {
	bit4 := publishPacket.TypeAndReserved - 16 - 32 // 去除 MQTT协议类型
	var needHandleRetain bool
	if bit4 >= 8 {         //重发标志 DUP, 0 表示第一次发这个消息
		bit4 -= 8
	}
	if bit4 %2 == 1{		 //保留标志 RETAIN ,为1 需要保存消息和服务等级
		needHandleRetain = true
		bit4 --
	}

	switch bit4 {		// 服务质量等级 QoS，左移1位越过retain
	case protocol.AtMostOnce  :
		// nothing to do
	case protocol.AtLeastOnce << 1:
		defer responsePubAck(conn, publishPacket.PacketIdentifier)
	case protocol.ExactOnce << 1:
		defer responsePubRec(conn,publishPacket.PacketIdentifier)
	case protocol.None << 1:
		// nothing to do
	}

	if needHandleRetain{
		protocolPacket.HandleRetain()
	}
	message := new(common.Message)
	fmt.Println(publishPacket.Payload)
	message = message.UnPack(publishPacket.Payload)
	fmt.Printf("主题 %s 生产了: %s ", publishPacket.TopicName,message.MsgKey )
	p.queue.Push(publishPacket.TopicName, *message)
	p.msgChan <- *message

}

func responsePubAck(conn net.Conn,identify uint16) {
	fmt.Println("发送puback")
	pubAckPacket := protocolPacket.NewPubAckPacket(identify)
	pubAckPacket.Write(conn)
	//data := utils.StructToBytes(pubAckPacket)
	//conn.Write(data)
}

func responsePubRec(conn net.Conn, identify uint16) {
	fmt.Println("准备返回pubRec")
	// PUBREC – 发布收到（QoS 2，第一步)
	pubRecPacket := protocolPacket.NewPubRecPacket(identify)
	pubRecPacket.Write(conn)


	// 等待 rel
	var fh protocolPacket.FixedHeader
	var pubRelPacket protocolPacket.PubRelPacket
	if err := fh.Read(conn);err !=nil{
		fmt.Println("接收pubRel包头内容错误",err)
	}
	pubRelPacket.Read(conn,fh)

	// PUBCOMP – 发布完成（QoS 2，第三步)
	fmt.Println("准备返回pubComp")
	pubCompPacket := protocolPacket.NewPubCompPacket(pubRelPacket.PacketIdentifier)
	pubCompPacket.Write(conn)
	// todo identify pool
	return
}