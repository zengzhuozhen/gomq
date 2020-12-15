package service

import (
	"github.com/urfave/cli/v2"
)

var (
	topic   string
	connect string
	qos     int
	retain  int
)

func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "topic",
			Usage:       "the topic you care",
			Destination: &topic,
		},
		&cli.StringFlag{
			Name:        "connect",
			Usage:       "connect the server broker",
			Destination: &connect,
		},
		&cli.IntFlag{
			Name:        "qos",
			Usage:       "set the message qos",
			Value:       0,
			Destination: &qos,
		},
		&cli.IntFlag{
			Name:        "retain",
			Usage:       "set the message retain",
			Value:       0,
			Destination: &retain,
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
			Name:   "list",
			Usage:  "list message",
			Action: ListMessage,
		},
		{
			Name:   "version",
			Usage:  "get current version",
			Action: GetVersion,
		},
	}
}
