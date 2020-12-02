package broker

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"gomq/common"
	"gomq/server/service"
	"gomq/server/store"
	"time"
)

type LeaderBroker struct {
	*Broker
}

func NewLeaderBroker(b *Broker) *LeaderBroker {
	return &LeaderBroker{b}
}

func (l *LeaderBroker) Run() {
	l.wg = errgroup.Group{}
	l.wg.Go(l.startConnLoop)
	l.wg.Go(l.startTcpServer)
	l.wg.Go(l.startHttpServer)
	l.wg.Go(l.startPersistent)
	l.wg.Go(l.handleSignal)
	l.wg.Go(l.MemberReceiver.Broadcast)
	_ = l.wg.Wait()
}

func (l *LeaderBroker) startPersistent() error {
	if l.opt.needPersistent == false{
		return  nil
	}
	fmt.Println("开启持久化协程")
	l.persistent = store.NewFileStore(l.opt.savePath)
	l.persistent.Open()
	l.persistent.Load()
	for {
		var data common.MessageUnit
		data = <-l.ProducerReceiver.Queue.PersistentChan
		fmt.Println("接收到持久化消息单元")

		l.persistent.Append(data)
		l.MemberReceiver.HP += 100 // 自己做了持久化，更新高水位线，基于内存的无效
		l.MemberReceiver.BroadcastChan <- data
	}
}

func (l *LeaderBroker) startConnLoop() error {
	fmt.Println("开启监听连接循环")
	for {
		activeConn := l.ConsumerReceiver.Pool.ForeachActiveConn()
		if len(activeConn) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, uid := range activeConn {
			topicList := l.ConsumerReceiver.Pool.Topic[uid]
			for k, topic := range topicList {
				position := l.ConsumerReceiver.Pool.Position[uid][k]
				if msg, err := l.ProducerReceiver.Queue.Pop(topic, position); err == nil {
					l.ConsumerReceiver.Pool.UpdatePosition(uid, topic)
					l.ConsumerReceiver.ChanAssemble[uid][k] <- msg
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}
}

func (l *LeaderBroker) startTcpServer() error {
	fmt.Println("开启tcp server...")
	tcp := service.NewTCP( l.opt.endPoint,l.ProducerReceiver, l.ConsumerReceiver, l.MemberReceiver)
	tcp.Start()
	return nil
}

func (l *LeaderBroker) startHttpServer() error {
	fmt.Println("开启http server... ")
	http := service.NewHTTP(l.ProducerReceiver,l.ConsumerReceiver)
	http.Start()
	return nil
}
