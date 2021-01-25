package main

import (
	"flag"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/server/broker"
	_ "net/http/pprof"
)


// todo 零拷贝 copyBuffer , impl writeTo and ReadFrom
func main() {
	flag.Parse()
	opts := broker.NewOption(broker.Leader, *common.Endpoint, *common.Dirname, []string{common.EtcdUrl})
	broker.NewBroker(opts).Run()
}
