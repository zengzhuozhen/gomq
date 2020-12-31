package broker

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"gomq/common"
	"gomq/server/service"
)

type LeaderBroker struct {
	*Broker
}

func NewLeaderBroker(b *Broker) *LeaderBroker {
	return &LeaderBroker{b}
}

func (l *LeaderBroker) Run() {
	l.wg = errgroup.Group{}
	l.wg.Go(l.startTcpServer)
	l.wg.Go(l.startHttpServer)
	l.wg.Go(l.startPersistent)
	l.wg.Go(l.handleSignal)
	l.wg.Go(l.MemberReceiver.Broadcast)
	_ = l.wg.Wait()
}

func (l *LeaderBroker) startPersistent() error {
	fmt.Println("开启持久化协程")
	for {
		var data common.MessageUnit
		data = <-l.ProducerReceiver.RetainQueue.ToPersistentChan
		fmt.Println("接收到持久化消息单元")
		l.persistent.Open(data.Topic)
		l.persistent.Append(data)
		l.MemberReceiver.HP += 100 // 自己做了持久化，更新高水位线，基于内存的无效
		l.MemberReceiver.BroadcastChan <- data
	}
}

func (l *LeaderBroker) startTcpServer() error {
	fmt.Println("开启tcp server...")
	tcp := service.NewTCP(l.opt.endPoint, l.ProducerReceiver, l.ConsumerReceiver, l.MemberReceiver)
	tcp.Start()
	return nil
}

func (l *LeaderBroker) startHttpServer() error {
	fmt.Println("开启http server... ")
	http := service.NewHTTP(l.ProducerReceiver, l.ConsumerReceiver)
	http.Start()
	return nil
}
