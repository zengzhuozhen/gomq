package broker

import (
	"bufio"
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/server/service"
	"github.com/zengzhuozhen/gomq/server/store"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

type LeaderBroker struct {
	*Broker
}

func NewLeaderBroker(b *Broker) *LeaderBroker {
	return &LeaderBroker{b}
}

func (l *LeaderBroker) Run() {
	l.run(l.startTcpServer, l.startHttpServer, l.startPersistent, l.startPprof, l.handleSignal, l.startBroadcast, l.startLogCompact)
}

func (l *LeaderBroker) startPersistent() error {
	log.Infof("开启持久化协程")
	for {
		var data common.MessageUnit
		data = <-l.ProducerReceiver.RetainQueue.ToPersistentChan
		log.Debugf("接收到持久化消息单元")
		l.persistent.Open(data.Topic)
		l.persistent.Append(data)
		l.MemberReceiver.HP++ // 自己做了持久化，更新高水位线，基于内存的无效
		l.MemberReceiver.BroadcastChan <- data
	}
}

func (l *LeaderBroker) startTcpServer() error {
	log.Infof("开启tcp server")
	tcpServer := service.NewTCP(l.opt.endPoint, l.ProducerReceiver, l.ConsumerReceiver, l.MemberReceiver)
	tcpServer.Start()
	return nil
}

func (l *LeaderBroker) startHttpServer() error {
	log.Infof("开启http server")
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
	log.Infof("开启broadcast")
	return l.MemberReceiver.Broadcast()
}

func (l *LeaderBroker) startLogCompact() error {
	log.Infof("开启日志压缩")
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			switch l.persistent.(type) {
			case *store.FileStore:
				fileStore := l.persistent.(*store.FileStore)
				topics := fileStore.GetAllTopics()
				log.Debugf("压缩日志的topic:", topics)
				for _, topic := range topics { // 相当于重新覆盖，考虑更好的实现
					i := 0
					var dirtyRow []int64
					msgMap := make(map[string]common.MessageUnitWithSort)
					messages := l.persistent.ReadAll(topic)
					for _, message := range messages {
						if msg, exist := msgMap[message.Data.MsgKey]; exist { // 从前往后扫描，出现重复的key，说明前面的key需要被清除
							dirtyRow = append(dirtyRow, msg.Sort)
						}
						msgMap[message.Data.MsgKey] = common.MessageUnitWithSort{
							Sort:        int64(i),
							MessageUnit: message,
						}
						i++
					}
					fd := fileStore.GetFd(topic)
					l.compact(fd, dirtyRow)
				}
			}
			return nil
		}
	}

}

// 日志压缩:逐行读取，写入buffer,最后重新覆盖源文件
func (l *LeaderBroker) compact(fd *os.File, dirtyRows []int64) {
	log.Debugf("开始清除日志行", dirtyRows)
	reader := bufio.NewReader(fd)
	res := make([]byte, 0)
	var i int64
	for {
		str, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if !common.FindInt64(i, dirtyRows) {
			res = append(res, []byte(str)...)
		}
		i++
	}
	fd.Truncate(0)
	fd.Write(res)
	fd.Sync()
	fd.Close()
}
