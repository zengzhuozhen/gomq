package main

import (
	"context"
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
	ctx := context.Background()
	consumerPool := consumer.NewPool()
	producerChannel := make(chan common.Message)
	consumerChannel := make(chan common.Message)

	producerReceiver := producer.NewProducerReceiver(producerChannel, queue)
	consumerReceiver := consumer.NewConsumerReceiver(consumerChannel, consumerPool)

	g := errgroup.Group{}
	g.Go(func() error {
		fmt.Println("监听发布....")
		for {
			select {
			case msg := <-producerChannel:
				fmt.Println("记录入队数据", msg.MsgKey)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	g.Go(func() error {
		fmt.Println("监听消费....")
		for {
			activeConn := consumerPool.ForeachActiveConn()
			if len(activeConn) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			for _, uid := range activeConn {
				topic := consumerPool.Topic[uid]
				position := consumerPool.Position[uid]
				if msg, err := queue.Pop(topic, position); err == nil {
					consumerPool.UpdatePosition(uid)
					consumerChannel <- msg
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	})

	g.Go(func() error {
		fmt.Println("开启tcp service...")
		listener := service.NewListener("tcp", ":9000", consumerPool)
		listener.Start(producerReceiver, consumerReceiver)
		return nil
	})

	g.Wait()

}
