package common

type ClientState struct {
	// 已经发送给服务端，但是还没有完成确认的QoS 1和QoS 2级别的消息
	Pub map[uint16]interface{}
	Rel map[uint16]interface{}
	// 已从服务端接收，但是还没有完成确认的QoS 2级别的消息
	Rec map[uint16]interface{}
}

type ServerState struct {
	// 客户端订阅消息
	SubscribeInfo []string
	// 已经发送给客户端，但是还没有完成确认的QoS 1和QoS 2级别的消息
	Pub map[uint16]interface{} // to consumer
	// 即将传输给客户端的QoS 1和QoS 2级别的消息
	Ack  map[uint16]interface{} // to producer
	Rel  map[uint16]interface{} // to producer and consumer
	Comp map[uint16]interface{} // to producer
	// 已从客户端接收，但是还没有完成确认的QoS 2级别的消息
	Rec map[uint16]interface{} //  to consumer
}


func NewServerState() *ServerState {
	return &ServerState{
		SubscribeInfo: []string{},
		Pub:           make(map[uint16]interface{}),
		Ack:           make(map[uint16]interface{}),
		Rel:           make(map[uint16]interface{}),
		Comp:          make(map[uint16]interface{}),
		Rec:           make(map[uint16]interface{}),
	}
}