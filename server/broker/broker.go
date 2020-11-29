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

type IBroker interface {
	Run()
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

func NewBroker(opt *option) IBroker {
	broker := new(Broker)
	broker.brokerId = uuid.New().String()
	broker.opt = opt
	broker.ProducerReceiver = service.NewProducerReceiver()
	broker.ConsumerReceiver = service.NewConsumerReceiver(make(map[string][]common.MsgUnitChan, 1024))
	broker.MemberReceiver = service.NewMemberReceiver()
	broker.FollowersRemote = make(map[string]string)

	broker.register()
	fmt.Println("初始化broker成功，ID:" + broker.brokerId)

	if broker.opt.identity == Leader {
		return NewLeaderBroker(broker)
	} else {
		return NewMemberBroker(broker)
	}

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

	fmt.Println("获取leader")
	resp, _ = kv.Get(ctx, FollowerPath, clientv3.WithPrefix())
	for _, i := range resp.Kvs {
		clientId := strings.SplitAfter(string(i.Key), FollowerPath)[1]
		b.FollowersRemote[clientId] = string(i.Value)
	}
}

func (b *Broker) startPprof() error {
	fmt.Println("开启pprof")
	ip := "127.0.0.1:6060"
	if err := http.ListenAndServe(ip, nil); err != nil {
		fmt.Printf("start pprof failed on %s\n", ip)
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
		_, err = kv.Delete(context.TODO(), fmt.Sprintf("%s/%s", FollowerPath, b.brokerId))
	}
	return err
}


