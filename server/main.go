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
	producerChannel := make(chan common.Message, 10)
	consumerChannel := make(chan common.Message, 10)

	producerReceiver := producer.NewProducerReceiver(producerChannel)
	consumerReceiver := consumer.NewConsumerReceiver(consumerChannel)

	g := errgroup.Group{}
	g.Go(func() error {
		fmt.Println("监听发布....")
		for {
			select {
			case msg := <-producerChannel:
				fmt.Println("记录入队数据")
				queue.Push(msg)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	g.Go(func() error {
		fmt.Println("监听消费....")
		for {
			if msg := queue.Pop(); msg != *new(common.Message) {
				fmt.Println("记录出队数据")
				// 判断是否有消费者连接，有才发送，没有则堆积数据
				consumerChannel <- msg
			}else{
				time.Sleep(100 * time.Millisecond)
			}
		}
	})

	g.Go(func() error {
		fmt.Println("开启tcp service...")
		listener := service.NewListener("tcp",":9000")
		listener.Start(producerReceiver, consumerReceiver)
		return nil
	})


	g.Wait()

}
