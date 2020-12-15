package common

import "container/list"


// todo 断线重启 状态恢复

type ProducerWaitingFor struct {
	Ack  list.List			// 已发送，还没有ack的消息 (QoS-1)
	Rec  list.List			// 已发送，还没有rec的消息 (QoS-2)
	Comp list.List			// 已接收，还没有comp的消息 (QoS-2)
}



