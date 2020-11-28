package service

import (
	"github.com/urfave/cli/v2"
)

var (
	topic   string
	connect string
)

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "topic",
			Aliases:     []string{"t"},
			Usage:       "the topic you care",
			Required:    true,
			Destination: &topic,
		},
		&cli.StringFlag{
			Name:        "connect",
			Aliases:     []string{"c"},
			Usage:       "connect the server broker",
			Required:    false,
			Destination: &connect,
		},
	}
}

func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "pub",
			Usage:  "publish message ",
			Action: PublishMessage,
		},
		{
			Name:   "sub",
			Usage:  "subscribe message",
			Action: SubscribeMessage,
		},
		{
			Name:    "list",
			Usage:   "list message",
			Action : ListTopic,
		},
	}
}
