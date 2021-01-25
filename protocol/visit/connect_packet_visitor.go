package visit

import (
	"context"
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/protocol/packet"
	"github.com/zengzhuozhen/gomq/protocol/utils"
	"go.etcd.io/etcd/clientv3"
	"reflect"
	"time"
)


type ConnectPacketVisitor struct {
	filterVisitor packet.Visitor
}

func (v *ConnectPacketVisitor) Visit(fn packet.VisitorFunc) error {
	return v.filterVisitor.Visit(fn)
}

func NewConnectPacketVisitor(visitor packet.Visitor) *ConnectPacketVisitor {
	return &ConnectPacketVisitor{
		filterVisitor: packet.NewFilteredVisitor(visitor,
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

func protocolNameValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	if reflect.DeepEqual(connectPacket.TypeAndReserved, utils.EncodeString("MQTT")) {
		return fmt.Errorf("客户端使用的协议错误")
	}
	return nil
}

func protocolLevelValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	if !connectPacket.IsSuitableProtocolLevel() {
		return fmt.Errorf("不满足客户端要求的协议等级")
		//todo 发送ack信息: responseConnectAck(conn, protocol.UnSupportProtocolVersion)
	}
	return nil
}

func handleConnectFlag(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	if !connectPacket.IsReserved() {
		return fmt.Errorf("CONNECT控制报文的保留标志位必须为0")
	}
	// todo
	//handler.handleCleanSession()
	//handler.handleWillFlag()
	//handler.handleWillQos()
	//handler.handleWillRetain()
	//handler.handleUserNameFlag()
	//handler.handlePasswordFlag()
	return nil
}

func handleKeepAlive(controlPacket packet.ControlPacket) error {
	// todo
	return nil
}

func clientIdentifierValidate(controlPacket packet.ControlPacket) error {
	connectPacket := controlPacket.(*packet.ConnectPacket)
	_, connectPayLoad := connectPacket.ProvisionConnectFlagsAndPayLoad()
	if !connectPayLoad.IsLegalClientId() {
		return fmt.Errorf("客户端唯一标识错误")
	}

	if !connectPayLoad.IsAuthorizedClient() {
		return fmt.Errorf("客户端未授权")
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

	etcdClient, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{common.EtcdUrl},
		DialTimeout: 10 * time.Second,
	})

	connectPacket := controlPacket.(*packet.ConnectPacket)
	_, connectPayLoad := connectPacket.ProvisionConnectFlagsAndPayLoad()

	getSecretFunc := func(username string) string {
		getResp, _ := etcdClient.KV.Get(context.TODO(), username)
		password = string(getResp.Kvs[0].Value)
		return password
	}

	if !connectPayLoad.IsCorrectSecret(getSecretFunc) {
		panic("客户端user和password错误")
	}
	return nil
}