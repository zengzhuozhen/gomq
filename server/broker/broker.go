package broker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zengzhuozhen/gomq/client"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/server/service"
	"github.com/zengzhuozhen/gomq/server/store"
	"go.etcd.io/etcd/clientv3"
	"golang.org/x/sync/errgroup"
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

type option struct {
	identity int
	endPoint string
	etcdUrls []string
	dirname  string
}

func NewOption(identity int, endPoint, dirname string, etcdUrls []string) *option {
	return &option{
		identity: identity,
		endPoint: endPoint,
		etcdUrls: etcdUrls,
		dirname:  dirname,
	}
}

type IBroker interface {
	Run()
}

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
	session          map[string]*common.ServerState

	memberClient *client.Member // 作为Member启动时持有的客户端
}

func NewBroker(opt *option) IBroker {
	broker := new(Broker)
	broker.brokerId = uuid.New().String()
	broker.opt = opt

	queue := common.NewQueue()
	broker.persistent = store.NewFileStore(broker.opt.dirname)
	broker.ProducerReceiver = service.NewProducerReceiver(queue, broker.persistent.ReadAll, broker.persistent.Reset, broker.persistent.Cap)
	broker.ConsumerReceiver = service.NewConsumerReceiver(make(map[string][]common.MsgUnitChan))
	broker.MemberReceiver = service.NewMemberReceiver(queue)
	broker.FollowersRemote = make(map[string]string)
	broker.session = make(map[string]*common.ServerState)

	broker.register()
	log.Debugf("初始化broker成功，ID:" + broker.brokerId)

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
	log.Debugf("Broker.HandleSignal receive signal %s \n", sig)
	err := b.gracefulStop()
	if err == nil {
		log.Debugf("Graceful Exit")
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
	b.persistent.Close()
	return err
}

func (b *Broker) run(fns ...func() error) {
	b.wg = errgroup.Group{}
	for _, fn := range fns {
		b.wg.Go(fn)
	}
	_ = b.wg.Wait()
}
