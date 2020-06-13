package protocol

const (
	Reserved int = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

// QoS
const (
	AtMostOnce = iota
	AtLeastOnce
	ExactOnce
	None
)

const (
//DUP1  重发标志 DUP
//QoS2  服务质量等级 QoS
//RETAIN3 保留标志 RETAIN
)
