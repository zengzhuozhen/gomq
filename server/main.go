package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"gomq/client"
	"gomq/common"
	"time"
)


func main() {

	queue := common.NewQueue()
	ctx := context.Background()
	producerChannel := make(chan common.Message, 10)
	consumerChannel := make(chan common.Message, 10)

	producer := client.NewProducer(producerChannel)
	consumer := client.NewConsumer(consumerChannel)

	g := errgroup.Group{}
	g.Go(func() error {
		for {
			select {
			case msg := <-producerChannel:
				queue.Push(msg)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	g.Go(func() error {
		fmt.Println("启动消费....")
		for {
			if msg := queue.Pop(); msg != *new(common.Message) {
				consumerChannel <- msg
			}else{
				time.Sleep(100 * time.Millisecond)
			}
		}
	})
	g.Go(func() error {
		consumer.Consume(ctx)
		return nil
	})

	msg := common.NewMessage(1, "hello", "hello world")
	producer.Produce(msg)
	producer.Produce(msg)
	producer.Produce(msg)

	g.Wait()

}
