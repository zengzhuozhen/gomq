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

type serverType int32

const (
	Leader serverType = 1
	Member serverType = 2
)

type option struct {
	identity serverType
	endPoint string
	etcdUrls []string
	dirname  string
}

type Option func(broker *Broker)

func ServerType(serverType serverType) Option {
	return func(broker *Broker) {
		broker.opt.identity = serverType
	}
}

func EndPoint(endpoint string) Option {
	return func(broker *Broker) {
		broker.opt.endPoint = endpoint
	}
}

func EtcdUrl(etcdUrl []string) Option {
	return func(broker *Broker) {
		broker.opt.etcdUrls = etcdUrl
	}
}

func Dirname(dirname string) Option {
	return func(broker *Broker) {
		broker.opt.dirname = dirname
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
	memberClient     *client.Member // 作为Member启动时持有的客户端
	ctx              context.Context
	cancelFunc       context.CancelFunc
}

func NewBroker(options ...Option) IBroker {
	broker := new(Broker)
	broker.opt = new(option)
	broker.brokerId = uuid.New().String()
	broker.ctx, broker.cancelFunc = context.WithCancel(context.Background())
	for _, option := range options {
		option(broker)
	}
	queue := common.NewQueue()
	broker.persistent = store.NewFileStore(broker.opt.dirname)
	broker.ProducerReceiver = service.NewProducerReceiver(queue, broker.persistent)
	broker.ConsumerReceiver = service.NewConsumerReceiver(make(map[string][]common.MsgUnitChan))
	broker.MemberReceiver = service.NewMemberReceiver(queue)
	broker.FollowersRemote = make(map[string]string)
	broker.register()
	log.Debugf("初始化broker成功，ID:" + broker.brokerId)

	if broker.opt.identity == Member {
		go broker.watchLeader()
		return NewMemberBroker(broker)
	}
	return NewLeaderBroker(broker)

}

func (b *Broker) register() {
	config := clientv3.Config{
		Endpoints:   b.opt.etcdUrls,
		DialTimeout: 10 * time.Second,
	}
	b.RegisterCenter, _ = clientv3.New(config)
	kv := clientv3.NewKV(b.RegisterCenter)
	kv.Txn(b.ctx).
		If(clientv3.Compare(clientv3.Version(LeaderPath), "=", 0)).
		Then(clientv3.OpPut(LeaderPath, b.opt.endPoint)).
		Else(clientv3.OpPut(fmt.Sprintf("%s%s", FollowerPath, b.brokerId), b.opt.endPoint)).
		Commit()
	b.updateIdentity(kv)
}

// 监听leader标志，当主节点宕机时，自动替换
// Leader和Member运行时的逻辑不一致，替换为主节点后需要将当前Broker转换为LeaderBroker并启动
func (b *Broker) watchLeader() {
	for {
		watchChan := b.RegisterCenter.Watch(b.ctx, LeaderPath)
		<-watchChan
		kv := clientv3.NewKV(b.RegisterCenter)
		kv.Txn(b.ctx).
			If(clientv3.Compare(clientv3.Version(LeaderPath), "=", 0)).
			Then(
				clientv3.OpPut(LeaderPath, b.opt.endPoint),
				clientv3.OpPut(LeaderId, b.brokerId),
				clientv3.OpDelete(fmt.Sprintf("%s%s", FollowerPath, b.brokerId)),
			).
			Commit()
		b.updateIdentity(kv)
		if b.opt.identity == Leader {
			goto LeaderRun
		}
	}
LeaderRun:
	leader := NewLeaderBroker(b)
	leader.Run()
}

// updateIdentity 更新当前broker的身份
func (b *Broker) updateIdentity(kv clientv3.KV) {
	// 获取leader地址
	resp, _ := kv.Get(b.ctx, LeaderPath)
	leaderRemote := string(resp.Kvs[0].Value)
	b.LeaderAddress = leaderRemote

	if leaderRemote == b.opt.endPoint {
		b.opt.identity = Leader
		_, _ = kv.Put(b.ctx, LeaderId, b.brokerId)
	} else {
		b.opt.identity = Member
	}
	// 获取leader ID
	resp, _ = kv.Get(b.ctx, LeaderId)
	b.LeaderId = string(resp.Kvs[0].Value)
	resp, _ = kv.Get(b.ctx, FollowerPath, clientv3.WithPrefix())
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
	b.cancelFunc()
	return err
}

func (b *Broker) run(fns ...func() error) {
	b.wg = errgroup.Group{}
	for _, fn := range fns {
		b.wg.Go(fn)
	}
	_ = b.wg.Wait()
}
