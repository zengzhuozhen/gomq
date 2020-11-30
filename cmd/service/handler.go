package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gomq/client"
	"gomq/cmd/do"
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

func ListTopic(context *cli.Context) error {
	resp := common.HttpGet(fmt.Sprintf("http://127.0.0.1:8000/topic/%s",context.Args().First()))
	listDo := new(do.MessagesDo)
	_ = json.Unmarshal([]byte(resp),listDo)
	fmt.Println("topic:",listDo.TopicName)
	for _ , msg := range listDo.MessageList{
		fmt.Println(msg)
	}
	return nil
}


func GetVersion(context *cli.Context) error {

	resp := common.HttpGet("http://127.0.0.1:8000/version")
	versionDo := new(do.VersionDo)
	_ = json.Unmarshal([]byte(resp),versionDo)
	fmt.Println("gomq version:",versionDo.Gomq)
	fmt.Println("gomqctl version:",versionDo.GomqCtl)
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
