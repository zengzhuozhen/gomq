package protocol

// 标准MQTT标识
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

// 自定义标识  255(byte最大值) > X > 240(MQTT协议封包最大值)
const(
	SYNCREQ byte = 250 + iota
	SYNCACK
	SYNCOFFSET
)

//  QoS 服务质量等级
const(
	AtMostOnce = iota
	AtLeastOnce
	ExactOnce
)

// 是否保留发送数据包
const (
	NotNeedRetain = 0
	NeedRetain = 1
)

// Connect包验证错误
const (
	ConnectAccess = iota
	UnSupportProtocolType
	UnSupportProtocolVersion
	UnSupportClientIdentity
	UnAvailableService
	UserAndPassError
	UnAuthorization
	ErrorParams
)
