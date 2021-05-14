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

const(
	// 自定义标识  255(byte最大值) > X > 240(MQTT协议封包最大值)
	SYNCREQ byte = 250 + iota
	SYNCACK
	SYNCOFFSET
)

const(
	// QoS 服务质量等级
	AtMostOnce = iota
	AtLeastOnce
	ExactOnce
)
const (
	// 是否保留发送数据包
	NotNeedRetain = 0
	NeedRetain = 1
)

const (
	// Connect包验证错误
	ConnectAccess = iota
	UnSupportProtocolType
	UnSupportProtocolVersion
	UnSupportClientIdentity
	UnAvailableService
	UserAndPassError
	UnAuthorization
	ErrorParams
)
