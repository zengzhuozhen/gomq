package visit

import (
	"context"
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/packet"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"go.etcd.io/etcd/clientv3"
	"reflect"
	"time"
)

type ConnectError struct {
	Code    int32
	Message string
}

func (e ConnectError) Error() string {
	return fmt.Sprintf(e.Message)
}

func NewConnectError(code int32, message string) error {
	return ConnectError{
		Code:    code,
		Message: message,
	}
}

type ConnectPacketVisitor struct {
	filteredVisitor packet.Visitor
}

func (v *ConnectPacketVisitor) Visit(fn packet.VisitorFunc) error {
	return v.filteredVisitor.Visit(fn)
}

func NewConnectPacketVisitor(visitor packet.Visitor, connection *common.ConnectionAbstract) *ConnectPacketVisitor {
	container := container{connection: connection}
	return &ConnectPacketVisitor{
		filteredVisitor: packet.NewFilteredVisitor(visitor,
			container.protocolNameValidate,
			container.protocolLevelValidate,
			container.handleConnectFlag,
			container.handleKeepAlive,
			container.clientIdentifierValidate,
			container.handleWillTopic,
			container.handleWillMessage,
			container.handleUserNameAndPassword,
		),
	}
}

type ConnectFlagVisitor struct {
	decoratedVisitor packet.Visitor
}

func (v *ConnectFlagVisitor) Visit(fn packet.VisitorFunc) error {
	return v.decoratedVisitor.Visit(fn)
}

func newConnectFlagVisitor(visitor packet.Visitor, connection *common.ConnectionAbstract) *ConnectFlagVisitor {
	container := container{connection: connection}
	return &ConnectFlagVisitor{
		decoratedVisitor: packet.NewDecoratedVisitor(visitor,
			container.handleCleanSession,
			container.handleWillFlag,
			container.handleWillQos,
			container.handleWillRetain,
			container.handleUsernameFlag,
			container.handlePasswordFlag,
		),
	}
}

func (c container) protocolNameValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	if reflect.DeepEqual(connectPacket.TypeAndReserved, utils.EncodeString("MQTT")) {
		return NewConnectError(protocol.UnSupportProtocolType, "客户端使用的协议错误")
	}
	return nil
}

func (c container) protocolLevelValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	if !connectPacket.IsSuitableProtocolLevel() {
		return NewConnectError(protocol.UnSupportProtocolVersion, "不满足客户端要求的协议等级")
	}
	return nil
}

func (c container) handleConnectFlag(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	return newConnectFlagVisitor(&PacketVisitor{Packet: connectPacket}, c.connection).Visit(func(controlPacket packet.ControlPacket) error {
		if !connectPacket.IsReserved() {
			return NewConnectError(protocol.UnAvailableService, "CONNECT控制报文的保留标志位必须为0")
		}
		return nil
	})
}

// handleKeepAlive 如果保持连接的值非零，并且服务端在一点五倍的保持连接时间内没有收到客户端的控制报文，它必须断开客户端的网络连接，认为网络连接已断开
func (c container) handleKeepAlive(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	c.connection.KeepAlive = connectPacket.KeepAlive
	return nil
}

func (c container) clientIdentifierValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	_, connectPayLoad := connectPacket.ProvisionConnectFlagsAndPayLoad()
	if !connectPayLoad.IsLegalClientId() {
		return NewConnectError(protocol.UnSupportClientIdentity, "客户端唯一标识错误")
	}
	if !connectPayLoad.IsAuthorizedClient() {
		return NewConnectError(protocol.UnAuthorization, "客户端未授权")
	}
	return nil
}

func (c container) handleWillTopic(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	_, connectPayLoad := connectPacket.ProvisionConnectFlagsAndPayLoad()
	if c.connection.WillFlag{
		c.connection.WillTopic = connectPayLoad.WillTopic
	}
	return nil
}

func (c container) handleWillMessage(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	_, connectPayLoad := connectPacket.ProvisionConnectFlagsAndPayLoad()
	if c.connection.WillFlag{
		c.connection.WillMessage = connectPayLoad.WillMessage
	}
	return nil
}

func (c container) handleUserNameAndPassword(controlPacket packet.ControlPacket) error {
	var password string

	etcd, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{common.EtcdUrl},
		DialTimeout: 10 * time.Second,
	})

	connectPacket := controlPacket.(*packet.ConnectPacket)
	_, connectPayLoad := connectPacket.ProvisionConnectFlagsAndPayLoad()

	getSecretFunc := func(username string) string {
		getResp, _ := etcd.KV.Get(context.TODO(), username)
		password = string(getResp.Kvs[0].Value)
		return password
	}

	if !connectPayLoad.IsCorrectSecret(getSecretFunc) {
		return NewConnectError(protocol.UserAndPassError, "客户端账号密码错误")
	}
	_ = etcd.Close()
	return nil
}

func (container) handleCleanSession(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	connectFlags := connectPacket.ProvisionConnectFlags()
	if connectFlags.CleanSession{
		// cleanSession todo 清除服务端的session
	}
	return nil
}

// handleWillFlag
// 遗嘱消息发布的条件，包括但不限于：
// 服务端检测到了一个I/O错误或者网络故障。
// 客户端在保持连接（Keep Alive）的时间内未能通讯。
// 客户端没有先发送DISCONNECT报文直接关闭了网络连接。
// 由于协议错误服务端关闭了网络连接。
func (c container) handleWillFlag(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	connectFlags := connectPacket.ProvisionConnectFlags()
	// 如果遗嘱标志被设置为1，连接标志中的Will QoS和Will Retain字段会被服务端用到，同时有效载荷中必须包含Will Topic和Will Message字段 [MQTT-3.1.2-9]。
	// 一旦被发布或者服务端收到了客户端发送的DISCONNECT报文，遗嘱消息就必须从存储的会话状态中移除 [MQTT-3.1.2-10]。
	// 如果遗嘱标志被设置为0，连接标志中的Will QoS和Will Retain字段必须设置为0，并且有效载荷中不能包含Will Topic和Will Message字段 [MQTT-3.1.2-11]。
	// 如果遗嘱标志被设置为0，网络连接断开时，不能发送遗嘱消息 [MQTT-3.1.2-12]。
	c.connection.WillFlag = connectFlags.WillFlag
	return nil
}

func (c container) handleWillQos(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	connectFlags := connectPacket.ProvisionConnectFlags()
	// 如果遗嘱标志被设置为0，遗嘱QoS也必须设置为0(0x00) [MQTT-3.1.2-13]。
	// 如果遗嘱标志被设置为1，遗嘱QoS的值可以等于0(0x00)，1(0x01)，2(0x02)。它的值不能等于3 [MQTT-3.1.2-14]。
	if connectFlags.WillFlag == false && connectFlags.WillQos != 0 {
		return NewConnectError(protocol.ErrorParams, "遗嘱标志为0，遗嘱QoS也必须为0")
	}
	c.connection.WillQos = connectFlags.WillQos
	return nil
}

func (c container) handleWillRetain(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	connectFlags := connectPacket.ProvisionConnectFlags()
	// 如果遗嘱标志被设置为0，遗嘱保留（Will Retain）标志也必须设置为0 [MQTT-3.1.2-15]。
	// 如果遗嘱保留被设置为0，服务端必须将遗嘱消息当作非保留消息发布 [MQTT-3.1.2-16]。
	// 如果遗嘱保留被设置为1，服务端必须将遗嘱消息当作保留消息发布 [MQTT-3.1.2-17]。
	if connectFlags.WillFlag == false && connectFlags.WillRetain == true {
		return NewConnectError(protocol.ErrorParams, "遗嘱标志为0，遗嘱保留也必须为0")
	}
	c.connection.WillRetain = connectFlags.WillRetain
	return nil
}

func (c container) handleUsernameFlag(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	connectFlags := connectPacket.ProvisionConnectFlags()
	// 如果用户名（User Name）标志被设置为0，有效载荷中不能包含用户名字段 [MQTT-3.1.2-18]。
	// 如果用户名（User Name）标志被设置为1，有效载荷中必须包含用户名字段 [MQTT-3.1.2-19]。
	c.connection.UserNameFlag = connectFlags.UserNameFlag
	return nil
}

func (c container) handlePasswordFlag(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	connectFlags := connectPacket.ProvisionConnectFlags()
	// 如果密码（Password）标志被设置为0，有效载荷中不能包含密码字段 [MQTT-3.1.2-20]。
	// 如果密码（Password）标志被设置为1，有效载荷中必须包含密码字段 [MQTT-3.1.2-21]。
	// 如果用户名标志被设置为0，密码标志也必须设置为0 [MQTT-3.1.2-22]。
	if connectFlags.UserNameFlag == false && connectFlags.PasswordFlag == true {
		return NewConnectError(protocol.ErrorParams, "用户名标志为0，密码标志也必须设置为0")
	}
	c.connection.PasswordFlag = connectFlags.PasswordFlag
	return nil
}
