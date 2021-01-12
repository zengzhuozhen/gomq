package handler

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"github.com/zengzhuozhen/gomq/protocol"
	"github.com/zengzhuozhen/gomq/protocol/packet"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"reflect"
)

type ConnectPacketHandler interface {
	HandleAll() error
	// 协议
	HandleProtocolName()
	// 协议级别
	HandleProtocolLevel()
	// 连接标识
	HandleConnectFlag()
	// 保持通讯
	HandleKeepAlive()
	// 客户端唯一标识
	HandleClientIdentifier()
	// 遗嘱主题
	HandleWillTopic()
	// 遗嘱消息
	HandleWillMessage()
	// 用户名和密码
	HandleUserNameAndPassword()
	// 解包过程需要处理的错误
	ErrorTypeToAck() byte
}

type ConnectPacketHandle struct {
	etcdClient     *clientv3.Client
	connectPacket  *packet.ConnectPacket
	connectFlags   *packet.ConnectFlags
	connectPayLoad *packet.ConnectPacketPayLoad
	errorTypeToAck byte
}

func (handler *ConnectPacketHandle) HandleAll() error {
	var err error
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("%d", r)
		}
	}()

	// 可变包头内容
	handler.HandleProtocolName()
	handler.HandleProtocolLevel()
	handler.HandleConnectFlag()
	handler.HandleKeepAlive()

	// 包体内容
	handler.HandleClientIdentifier()
	handler.HandleWillTopic()
	handler.HandleWillMessage()
	handler.HandleUserNameAndPassword()

	return nil
}

func (handler *ConnectPacketHandle) HandleProtocolName() {
	if reflect.DeepEqual(handler.connectPacket.TypeAndReserved, utils.EncodeString("MQTT")) {
		panic("客户端使用的协议错误")
	}
}

func (handler *ConnectPacketHandle) HandleProtocolLevel() {
	if !handler.connectPacket.IsSuitableProtocolLevel() {
		handler.errorTypeToAck = protocol.UnSupportProtocolVersion
		panic("不满足客户端要求的协议等级")
		//todo 发送ack信息: responseConnectAck(conn, protocol.UnSupportProtocolVersion)
	}
}

func (handler *ConnectPacketHandle) HandleConnectFlag() {
	if !handler.connectPacket.IsReserved() {
		panic("CONNECT控制报文的保留标志位必须为0")
	}
	//handler.handleCleanSession()
	//handler.handleWillFlag()
	//handler.handleWillQos()
	//handler.handleWillRetain()
	//handler.handleUserNameFlag()
	//handler.handlePasswordFlag()
}

func (handler *ConnectPacketHandle) handleCleanSession() {
	panic("implement me")
}

func (handler *ConnectPacketHandle) handleWillFlag() {
	panic("implement me")
}

func (handler *ConnectPacketHandle) handleWillQos() {
	panic("implement me")
}

func (handler *ConnectPacketHandle) handleWillRetain() {
	panic("implement me")
}

func (handler *ConnectPacketHandle) handleUserNameFlag() {
	panic("implement me")
}

func (handler *ConnectPacketHandle) handlePasswordFlag() {
	panic("implement me")
}

func (handler *ConnectPacketHandle) HandleKeepAlive() {
	panic("implement me")
}

func (handler *ConnectPacketHandle) HandleClientIdentifier() {
	if !handler.connectPayLoad.IsLegalIdentifier() {
		handler.errorTypeToAck = protocol.UnSupportClientIdentity
		panic("客户端唯一标识错误")
	}

	if !handler.connectPayLoad.IsAuthorizedClient() {
		handler.errorTypeToAck = protocol.UnAuthorization
		panic("客户端未授权")
	}
}

func (handler *ConnectPacketHandle) HandleWillTopic() {

}

func (handler *ConnectPacketHandle) HandleWillMessage() {

}

func (handler *ConnectPacketHandle) HandleUserNameAndPassword() {
	var username, password string
	getResp, _ := handler.etcdClient.KV.Get(context.TODO(), username)
	username = string(getResp.Kvs[0].Key)
	password = string(getResp.Kvs[0].Value)

	if !handler.connectPayLoad.IsCorrectSecret(username, password) {
		handler.errorTypeToAck = protocol.UserAndPassError
		panic("客户端user和password错误")
	}
}

func (handler *ConnectPacketHandle) ErrorTypeToAck() byte {
	return handler.errorTypeToAck
}

func NewConnectPacketHandle(connectPacket *packet.ConnectPacket, connectFlags *packet.ConnectFlags, payLoad *packet.ConnectPacketPayLoad,client *clientv3.Client) ConnectPacketHandler {
	return &ConnectPacketHandle{
		etcdClient:     client,
		connectPacket:  connectPacket,
		connectFlags:   connectFlags,
		connectPayLoad: payLoad,
	}
}
