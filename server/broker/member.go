package broker

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/client"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/server/store"
)

type MemberBroker struct {
	*Broker
}

func NewMemberBroker(b *Broker) *MemberBroker {
	return &MemberBroker{b}
}

func (m MemberBroker) Run() {
	m.run(m.startSendSync, m.startPersistent, m.handleSignal)
}

func (m *MemberBroker) startSendSync() error {
	m.memberClient = client.NewMember(&client.Option{
		Protocol: "tcp",
		Address:  m.LeaderAddress,
		KeepAlive:  3,
	})
	return m.memberClient.SendSync()
}

func (m *MemberBroker) startPersistent() error {
	fmt.Println("开启持久化协程")
	m.persistent = store.NewMemStore()
	m.persistent.Open("")
	m.persistent.ReadAll("")
	for {
		var data common.MessageUnit
		data = <-m.memberClient.PersistentChan
		fmt.Println("同步Leader消息")
		m.persistent.Append(data)
	}
}
