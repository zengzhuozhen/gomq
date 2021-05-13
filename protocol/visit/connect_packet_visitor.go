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

func NewConnectPacketVisitor(visitor packet.Visitor) *ConnectPacketVisitor {
	return &ConnectPacketVisitor{
		filteredVisitor: packet.NewFilteredVisitor(visitor,
			protocolNameValidate,
			protocolLevelValidate,
			handleConnectFlag,
			handleKeepAlive,
			clientIdentifierValidate,
			handleWillTopic,
			handleWillMessage,
			handleUserNameAndPassword,
		),
	}
}

type ConnectFlagVisitor struct {
	decoratedVisitor packet.Visitor
}

func (v *ConnectFlagVisitor) Visit(fn packet.VisitorFunc) error {
	return v.decoratedVisitor.Visit(fn)
}

func newConnectFlagVisitor(visitor packet.Visitor) *ConnectFlagVisitor {
	return &ConnectFlagVisitor{
		decoratedVisitor: packet.NewDecoratedVisitor(visitor,
			handleCleanSession,
			handleWillFlag,
			handleWillQos,
			handleWillRetain,
			handleUsernameFlag,
			handlePasswordFlag,
		),
	}
}

func protocolNameValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	if reflect.DeepEqual(connectPacket.TypeAndReserved, utils.EncodeString("MQTT")) {
		return NewConnectError(protocol.UnSupportProtocolType, "客户端使用的协议错误")
	}
	return nil
}

func protocolLevelValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	if !connectPacket.IsSuitableProtocolLevel() {
		return NewConnectError(protocol.UnSupportProtocolVersion, "不满足客户端要求的协议等级")
	}
	return nil
}

func handleConnectFlag(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	return newConnectFlagVisitor(&PacketVisitor{Packet: connectPacket}).Visit(func(controlPacket packet.ControlPacket) error {
		if !connectPacket.IsReserved() {
			return NewConnectError(protocol.UnAvailableService,"CONNECT控制报文的保留标志位必须为0")
		}
		return nil
	})
}

// handleKeepAlive 如果保持连接的值非零，并且服务端在一点五倍的保持连接时间内没有收到客户端的控制报文，它必须断开客户端的网络连接，认为网络连接已断开
func handleKeepAlive(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	common.KeepAlice =  connectPacket.KeepAlive
	return nil
}

func clientIdentifierValidate(controlPacket packet.ControlPacket) error {
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

func handleWillTopic(controlPacket packet.ControlPacket) error {
	// todo
	return nil
}

func handleWillMessage(controlPacket packet.ControlPacket) error {
	// todo
	return nil
}

func handleUserNameAndPassword(controlPacket packet.ControlPacket) error {
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
	return nil
}

func handleCleanSession(controlPacket packet.ControlPacket) error {
	return nil
}

func handleWillFlag(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	connectFlags, _ := connectPacket.ProvisionConnectFlagsAndPayLoad()
	if connectFlags.WillFlag {
		// todo 设置遗嘱消息，服务端与客户端断开连接时发送该消息
	}
	return nil
}

func handleWillQos(controlPacket packet.ControlPacket) error {
	return nil
}

func handleWillRetain(controlPacket packet.ControlPacket) error {
	return nil
}

func handleUsernameFlag(controlPacket packet.ControlPacket) error {
	return nil
}

func handlePasswordFlag(controlPacket packet.ControlPacket) error {
	return nil
}
