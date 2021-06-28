package broker

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/client"
	"github.com/zengzhuozhen/gomq/common"
)

type MemberBroker struct {
	*Broker
}

func NewMemberBroker(b *Broker) *MemberBroker {
	m := &MemberBroker{b}
	m.memberClient = client.NewMember(&client.Option{
		Protocol: "tcp",
		Address:  m.LeaderAddress,
		KeepAlive:  30,
	})
	return m
}

func (m MemberBroker) Run() {
	m.run(m.startSendSync, m.startPersistent, m.handleSignal)
}

func (m *MemberBroker) startSendSync() error {
	return m.memberClient.SendSync()
}

func (m *MemberBroker) startPersistent() error {
	fmt.Println("开启持久化协程")
	for {
		var data common.MessageUnit
		data = <-m.memberClient.PersistentChan
		fmt.Println("同步Leader消息")
		m.persistent.Open(data.Topic)
		m.persistent.Append(data)
	}
}
