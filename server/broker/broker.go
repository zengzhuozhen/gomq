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

const LeaderId = "/service/mq/broker_leader_id"
const LeaderPath = "/services/mq/broker_leader"
const FollowerPath = "/services/mq/broker_follower/"

const (
	Leader = iota
	Member
)

const (
	StoreByMem = iota
	StoreByFile
)

type option struct {
	identity int
	endPoint string
	etcdUrls []string
}

func NewOption(identity int, endPoint string, etcdUrls []string) *option {
	return &option{
		identity: identity,
		endPoint: endPoint,
		etcdUrls: etcdUrls,
	}
}

// todo Broker 之上在加一层Leader和Member
type Broker struct {
	brokerId         string
	opt              *option
	wg               errgroup.Group
	ProducerReceiver *service.ProducerReceiver
	ConsumerReceiver *service.ConsumerReceiver
	MemberReceiver   *service.MemberReceiver
	persistent       store.Store
	LeaderId         string
	LeaderAddress    string
	FollowersRemote  map[string]string // clientId : ipAddress
	RegisterCenter   *clientv3.Client
	memberClient     *client.Member // 作为Member启动时持有的客户端
}

func NewBroker(opt *option) *Broker {
	broker := new(Broker)
	broker.brokerId = uuid.New().String()
	broker.opt = opt
	broker.ProducerReceiver = service.NewProducerReceiver()
	broker.ConsumerReceiver = service.NewConsumerReceiver(make(map[string][]common.MsgUnitChan, 1024))
	broker.MemberReceiver = service.NewMemberReceiver()
	broker.FollowersRemote = make(map[string]string)

	broker.register()

	fmt.Println("初始化broker成功，ID:" + broker.brokerId)
	return broker
}

func (b *Broker) Run() {
	if b.opt.identity == Member {
		b.runMember()
	} else {
		b.wg = errgroup.Group{}
		b.wg.Go(b.startPersistent)
		b.wg.Go(b.startConnLoop)
		b.wg.Go(b.startTcpServer)
		b.wg.Go(b.handleSignal)
		_ = b.wg.Wait()
	}
}

func (b *Broker) startPersistent() error {
	fmt.Println("开启持久化协程")
	b.persistent = store.NewFileStore()
	b.persistent.Open()
	b.persistent.Load()
	for {
		if b.opt.identity == Leader {
			data := <-b.ProducerReceiver.Queue.PersistentChan
			fmt.Println("接收到持久化消息单元")
			b.persistent.Append(data)
			if b.persistent.Cap()%100 == 0 { // 每100个元素做一次快照
				b.persistent.SnapShot()
			}
		} else {
			data := <-b.memberClient.PersistentChan
			fmt.Println("同步Leader消息")
			b.persistent.Append(data)
			if b.persistent.Cap()%100 == 0 { // 每100个元素做一次快照
				b.persistent.SnapShot()
			}

		}
	}
}

func (b *Broker) startConnLoop() error {
	fmt.Println("开启监听连接循环")
	for {
		activeConn := b.ConsumerReceiver.Pool.ForeachActiveConn()
		if len(activeConn) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, uid := range activeConn {
			topicList := b.ConsumerReceiver.Pool.Topic[uid]
			for k, topic := range topicList {
				position := b.ConsumerReceiver.Pool.Position[uid][k]
				if msg, err := b.ProducerReceiver.Queue.Pop(topic, position); err == nil {
					b.ConsumerReceiver.Pool.UpdatePosition(uid, topic)
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
	listener := service.NewListener("tcp", b.opt.endPoint, b.ProducerReceiver, b.ConsumerReceiver, b.MemberReceiver)
	listener.Start()
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
		Else(clientv3.OpPut(fmt.Sprintf("%s%s", FollowerPath, b.brokerId), b.opt.endPoint)).
		Commit()
	// 获取leader地址
	resp, _ := kv.Get(ctx, LeaderPath)
	leaderRemote := string(resp.Kvs[0].Value)
	b.LeaderAddress = leaderRemote

	if leaderRemote == b.opt.endPoint {
		b.opt.identity = Leader
		_, _ = kv.Put(ctx, LeaderId, b.brokerId)
	} else {
		// 需要判断是否为活跃节点，否则将替换为当前节点为leader,并修改etcd中的leader path
		b.opt.identity = Member
	}
	// 获取leader ID
	resp, _ = kv.Get(ctx, LeaderId)
	b.LeaderId = string(resp.Kvs[0].Value)

	resp, _ = kv.Get(ctx, FollowerPath, clientv3.WithPrefix())
	for _, i := range resp.Kvs {
		clientId := strings.SplitAfter(string(i.Key), FollowerPath)[1]
		b.FollowersRemote[clientId] = string(i.Value)
	}
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
		_, err = kv.Delete(context.TODO(), fmt.Sprintf("%s/%s", FollowerPath, b.brokerId))
	}
	return err
}

func (b *Broker) runMember() {
	// Member启动后连接Leader，注册信息，并同步数据
	// Member定时发送心跳，并告知同步情况
	// todo
	// Leader维持一个低水位，保证该水位之下的消息高可用，当低水位距离最新信息偏移太远时，Leader拒绝接收新的数据，通过Member同步来提高低水位线
	// Leader维持一个高水位，保证该水位之下的消息在当前节点可用，要保证Leader数据可靠，需要每次写入信息后刷盘，默认情况下高水位线与当前最新消息偏移一致，需要提高
	// 写入性能时，可以修改持久化策略，使高水位线在最新消息偏移之下

	host := strings.Split(b.LeaderAddress, ":")[0]
	port, _ := strconv.Atoi(strings.Split(b.LeaderAddress, ":")[1])
	b.memberClient = client.NewMember(&client.Option{
		Protocol: "tcp",
		Host:     host,
		Port:     port,
		Timeout:  3,
	})
	b.wg = errgroup.Group{}
	b.wg.Go(b.memberClient.SendSync)
	b.wg.Go(b.startPersistent)
	b.wg.Go(b.handleSignal)
	b.wg.Wait()
}
