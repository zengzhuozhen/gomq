package broker

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/server/service"
	"net/http"
	_ "net/http/pprof"
)

type LeaderBroker struct {
	*Broker
}

func NewLeaderBroker(b *Broker) *LeaderBroker {
	return &LeaderBroker{b}
}

func (l *LeaderBroker) Run() {
	l.run(l.startTcpServer, l.startHttpServer, l.startPersistent, l.startPprof, l.handleSignal, l.startBroadcast)
}

func (l *LeaderBroker) startPersistent() error {
	log.Infof("开启持久化协程")
	for {
		var data common.MessageUnit
		data = <-l.ProducerReceiver.RetainQueue.ToPersistentChan
		log.Debugf("接收到持久化消息单元")
		l.persistent.Open(data.Topic)
		l.persistent.Append(data)
		l.MemberReceiver.HP ++ // 自己做了持久化，更新高水位线，基于内存的无效
		l.MemberReceiver.BroadcastChan <- data
	}
}

func (l *LeaderBroker) startTcpServer() error {
	log.Infof("开启tcp server...")
	tcpServer := service.NewTCP(l.opt.endPoint, l.ProducerReceiver, l.ConsumerReceiver, l.MemberReceiver, l.RegisterCenter,l.session)
	tcpServer.Start()
	return nil
}

func (l *LeaderBroker) startHttpServer() error {
	log.Infof("开启http server... ")
	httpServer := service.NewHTTP(l.ProducerReceiver, l.ConsumerReceiver)
	httpServer.Start()
	return nil
}

func (l *LeaderBroker) startPprof() error {
	log.Infof("开启pprof")
	ip := "127.0.0.1:6060"
	if err := http.ListenAndServe(ip, nil); err != nil {
		fmt.Printf("start pprof failed on %s\n", ip)
	}
	return nil
}

func (l *LeaderBroker) startBroadcast() error {
	log.Infof("开启broadcast...")
	return l.MemberReceiver.Broadcast()
}