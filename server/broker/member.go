package broker

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"gomq/client"
	"gomq/common"
	"gomq/server/store"
)

type MemberBroker struct {
	*Broker
}

func NewMemberBroker(b *Broker) *MemberBroker {
	return &MemberBroker{b}
}

func (m MemberBroker) Run() {
	m.memberClient = client.NewMember(&client.Option{
		Protocol: "tcp",
		Address:  m.LeaderAddress,
		Timeout:  3,
	})
	m.wg = errgroup.Group{}
	m.wg.Go(m.memberClient.SendSync)
	m.wg.Go(m.startPersistent)
	m.wg.Go(m.handleSignal)
	m.wg.Wait()
}

func (m *MemberBroker) startPersistent() error {
	fmt.Println("开启持久化协程")
	m.persistent = store.NewFileStore()
	m.persistent.Open()
	m.persistent.Load()
	for {
		var data common.MessageUnit
		data = <-m.memberClient.PersistentChan
		fmt.Println("同步Leader消息")
		m.persistent.Append(data)
		if m.persistent.Cap()%100 == 0 { // 每100个元素做一次快照
			m.persistent.SnapShot()
		}
	}
}
