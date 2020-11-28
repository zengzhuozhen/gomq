package service

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gomq/client"
	"gomq/common"
	"strings"
)

func PublishMessage(c *cli.Context) error {
	producer := client.NewProducer(defaultOpts())
	body := strings.Join(c.Args().Slice(), " ")
	message := common.MessageUnit{
		Topic: topic,
		Data: common.Message{
			MsgKey: uuid.New().String(),
			Body:   body,
		},
	}
	producer.Publish(message, 0)
	fmt.Printf("publish a message %s to %s \n", body, topic)
	return nil
}

func SubscribeMessage(c *cli.Context) error {
	consumer := client.NewConsumer(defaultOpts())
	retChan := consumer.Subscribe([]string{topic})
	for msg := range retChan {
		fmt.Println(msg.Data.Body)
	}
	fmt.Printf("subscribe a message %s to %s \n", c.Args().First(), topic)
	return nil
}

func ListTopic(c *cli.Context) error {
	// todo http connect client to get server metadata
	return nil
}

func defaultOpts() *client.Option {
	return &client.Option{
		Protocol: "tcp",
		Address:  connect,
		Timeout:  3,
	}
}

func defaultHttpOpts() *client.Option {
	return &client.Option{
		Protocol: "http",
		Address:  connect,
		Timeout:  3,
	}
}
