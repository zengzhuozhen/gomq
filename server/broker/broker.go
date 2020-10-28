package broker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/sync/errgroup"
	"gomq/client"
	"gomq/common"
	"gomq/server/service"
	"gomq/server/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Identity int8

const LeaderPath = "/services/mq/broker_leader"
const FollowerPath = "/services/mq/broker_follower"

const (
	Leader   Identity = 0
	Follower Identity = 1
)

const (
	StoreByMem = iota
	StoreByFile
)

type option struct {
	identity Identity
	endPoint string
	etcdUrls []string
}

func NewOption(identity Identity, endPoint string, etcdUrls []string) *option {
	return &option{
		identity: identity,
		endPoint: endPoint,
		etcdUrls: etcdUrls,
	}
}

type Broker struct {
	brokerId         string
	opt              *option
	wg               errgroup.Group
	queue            *common.Queue
	ProducerReceiver *service.ProducerReceiver
	ConsumerReceiver *service.ConsumerReceiver
	ConnectPool      *service.Pool
	persistent       store.Store
	LeaderRemote     string
	FollowersRemote  map[string]string
	RegisterCenter   *clientv3.Client
}

func NewBroker(opt *option) *Broker {
	broker := new(Broker)
	broker.brokerId = uuid.New().String()
	broker.opt = opt
	broker.queue = common.NewQueue()
	broker.ConnectPool = service.NewPool()
	broker.ProducerReceiver = service.NewProducerReceiver(broker.queue)
	broker.ConsumerReceiver = service.NewConsumerReceiver(make(map[string][]common.MsgChan, 1024), broker.ConnectPool)
	broker.FollowersRemote = make(map[string]string)

	broker.register()
	fmt.Println("初始化broker成功，ID:" + broker.brokerId)
	return broker
}

func (b *Broker) Run() {
	b.wg = errgroup.Group{}

	b.wg.Go(b.startPersistent)
	b.wg.Go(b.startConnLoop)
	b.wg.Go(b.startTcpServer)
	b.wg.Go(b.startMemberSync)
	b.wg.Go(b.startPprof)
	b.wg.Go(b.handleSignal)

	_ = b.wg.Wait()
}

func (b *Broker) startPersistent() error {
	fmt.Println("开启持久化协程")
	b.persistent = store.NewFileStore()
	b.persistent.Open()
	b.persistent.Load()
	for data := range b.queue.PersistentChan {
		fmt.Println("接收到持久化信息")
		b.persistent.Append(data)
		if b.persistent.Cap()/100 == 0 { // 每100个元素做一次快照
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
	listener := service.NewListener("tcp", b.opt.endPoint)
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

func (b *Broker) register() {
	config := clientv3.Config{
		Endpoints:   b.opt.etcdUrls,
		DialTimeout: 10 * time.Second,
	}
	b.RegisterCenter, _ = clientv3.New(config)
	kv := clientv3.NewKV(b.RegisterCenter)
	ctx := context.Background()
	kv.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(LeaderPath), "=", 0)).
		Then(clientv3.OpPut(LeaderPath, b.opt.endPoint)).
		Else(clientv3.OpPut(fmt.Sprintf("%s/%s",FollowerPath,b.brokerId), b.opt.endPoint)).
		Commit()
	// 获取leader地址
	resp, _ := kv.Get(ctx, LeaderPath)
	leaderRemote := string(resp.Kvs[0].Value)

	if leaderRemote == b.opt.endPoint {
		b.opt.identity = Leader
	} else {
		b.opt.identity = Follower
	}

	resp, _ = kv.Get(ctx, FollowerPath+"/")
	for _, i := range resp.Kvs {
		b.FollowersRemote[string(i.Key)] = string(i.Value)
	}
}

func (b *Broker) startMemberSync() error {
	fmt.Println("开启集群同步")

	if len(b.FollowersRemote) < 1 {
		return nil
	}
	for data := range b.queue.MembersSyncChan {
		for _, member := range b.FollowersRemote {
			host := strings.Split(member, ":")[0]
			port, _ := strconv.Atoi(strings.Split(member, ":")[1])
			producer := client.NewProducer(&client.Option{
				Protocol: "tcp",
				Host:     host,
				Port:     port,
				Timeout:  3,
			})
			producer.Publish(data.Topic, data.Data, 1)
		}

	}
	return nil
}

func (b *Broker) handleSignal() error {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Broker.HandleSignal receive signal %s \n", sig)
	err := b.gracefulStop()
	if err == nil {
		fmt.Println("Graceful Exit")
		os.Exit(0)
	}
	return err

}

func (b *Broker) gracefulStop() error {
	var err error
	kv := clientv3.NewKV(b.RegisterCenter)
	if b.opt.identity == Leader {
		_, err = kv.Delete(context.TODO(), LeaderPath)
	} else {
		_, err = kv.Delete(context.TODO(), fmt.Sprintf("%s/%s",FollowerPath,b.brokerId))
	}
	return err
}
