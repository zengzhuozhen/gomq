package main

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"gomq/common"
	"gomq/server/consumer"
	"gomq/server/producer"
	"gomq/server/service"
	"time"
)


func main() {
	queue := common.NewQueue()
	consumerPool := consumer.NewPool()
	consumerChanAssemble := make(map[string][]common.MsgChan, 1024)
	producerReceiver := producer.NewProducerReceiver(queue)
	consumerReceiver := consumer.NewConsumerReceiver(consumerChanAssemble, consumerPool)

	g := errgroup.Group{}

	g.Go(func() error {
		fmt.Println("监听消费....")
		for {
			activeConn := consumerPool.ForeachActiveConn()
			if len(activeConn) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			for _, uid := range activeConn {
				topicList := consumerPool.Topic[uid]
				for k, topic := range topicList {
					position := consumerPool.Position[uid][k]
					if msg, err := queue.Pop(topic, position); err == nil {
						consumerPool.UpdatePosition(uid, topic)
						consumerReceiver.ChanAssemble[uid][k] <- msg
					} else {
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		}
	})

	g.Go(func() error {
		fmt.Println("开启tcp server...")
		listener := service.NewListener("tcp", ":9000")
		listener.Start(producerReceiver, consumerReceiver)
		return nil
	})

	g.Wait()

}
