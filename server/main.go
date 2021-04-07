package main

import (
	"flag"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/server/broker"
	_ "net/http/pprof"
)

func main() {
	flag.Parse()
	broker.NewBroker(
		broker.ServerType(broker.Leader),
		broker.EndPoint(*common.Endpoint),
		broker.Dirname(*common.Dirname),
		broker.EtcdUrl([]string{common.EtcdUrl})).
		Run()
}
