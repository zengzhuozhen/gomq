package main

import (
	"flag"
	"gomq/common"
	"gomq/server/broker"
	_ "net/http/pprof"
)

var endpoint = flag.String("endpoint", "127.0.0.1:9000", "mq服务运行地址")
var etcdurl string

func init(){
	if common.IsRunningInDocker() {
		etcdurl = "http://etcd:2379"
	}else {
		etcdurl = "127.0.0.1:2379"
	}
}


func main() {
	flag.Parse()
	opts := broker.NewOption(broker.Leader, *endpoint, []string{etcdurl})
	broker.NewBroker(opts).Run()
}
