package main

import (
	"flag"
	"gomq/server/broker"
	_ "net/http/pprof"
)


var endpoint = flag.String("p","127.0.0.1:9000","服务运行地址")

func main() {
	flag.Parse()
	opts := broker.NewOption(broker.Leader,*endpoint,[]string{"127.0.0.1:2379"})
	broker.NewBroker(opts).Run()
}
