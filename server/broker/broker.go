package broker

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"gomq/common"
	"gomq/server/service"
	"gomq/server/store"
	"net/http"
	"time"
)

const (
	Leader   = 0
	Follower = 1
)

const (
	StoreByMem = iota
	StoreByFile
)

type RegisterCenter struct {
	Address string
}

type Broker struct {
	wg               errgroup.Group
	queue            *common.Queue
	RegisterCenter   *RegisterCenter
	ProducerReceiver *service.ProducerReceiver
	ConsumerReceiver *service.ConsumerReceiver
	ConnectPool      *Pool
	persistent       store.Store
}

func NewBroker(identity int32) *Broker {
	connectPool := NewPool()
	broker := new(Broker)
	broker.queue = common.NewQueue()
	broker.ProducerReceiver = service.NewProducerReceiver(broker.queue)
	broker.ConsumerReceiver = service.NewConsumerReceiver(make(map[string][]common.MsgChan, 1024), connectPool)

	broker.wg = errgroup.Group{}

	broker.wg.Go(broker.startPersistent)
	broker.wg.Go(broker.startConnLoop)
	broker.wg.Go(broker.startTcpServer)
	broker.wg.Go(broker.startPprof)

	return broker
}

func (b *Broker) startPersistent() error {
	fmt.Println("开启持久化协程")
	b.persistent = store.NewMemStore()
	b.persistent.Open()
	b.persistent.Load()
	for data := range b.queue.PersistentChan {
		b.persistent.Append(data)
		if b.persistent.Cap() / 100 == 0 {  // 每100个元素做一次快照
			b.persistent.SnapShot()
		}
	}
	return nil
}

func (b *Broker) startConnLoop() error {
	fmt.Println("开启监听连接循环")
	for {
		activeConn := b.ConnectPool.ForeachActiveConn()
		if len(activeConn) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, uid := range activeConn {
			topicList := b.ConnectPool.Topic[uid]
			for k, topic := range topicList {
				position := b.ConnectPool.Position[uid][k]
				if msg, err := b.queue.Pop(topic, position); err == nil {
					b.ConnectPool.UpdatePosition(uid, topic)
					b.ConsumerReceiver.ChanAssemble[uid][k] <- msg
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}
}

func (b *Broker) startTcpServer() error {
	fmt.Println("开启tcp server...")
	listener := service.NewListener("tcp", ":9000")
	listener.Start(b.ProducerReceiver, b.ConsumerReceiver)
	return nil
}

func (b *Broker) startPprof() error {
	fmt.Println("开启pprof")
	ip := "127.0.0.1:6060"
	if err := http.ListenAndServe(ip, nil); err != nil {
		fmt.Printf("start pprof failed on %s\n", ip)
	}
	return nil
}

func (b *Broker) Run() {
	_ = b.wg.Wait()
}
