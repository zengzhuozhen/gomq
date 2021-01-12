package broker

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/client"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/server/store"
	"golang.org/x/sync/errgroup"
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
