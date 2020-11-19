package protocol

const (
	Reserved int = iota
	// 标准MQTT标识
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

const (
	// 自定义标识  255(byte最大值) > X > 240(MQTT协议封包最大值)
	SYNCREQ	 byte = 250 + iota
	SYNCACK
	SYNCOFFSET
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
