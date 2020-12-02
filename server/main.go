package main

import (
	"flag"
	"gomq/common"
	"gomq/server/broker"
	_ "net/http/pprof"
)

var endpoint = flag.String("endpoint", "0.0.0.0:9000", "mq服务运行地址")
var savepath = flag.String("savepath", "/var/log/tempmq.log", "mq数据保存路径")
var persistent = flag.Bool("persistent", true, "是否开启持久化")
var etcdurl string

func init() {
	if common.IsRunningInDocker() {
		etcdurl = "http://etcd:2379"
	} else {
		etcdurl = "127.0.0.1:2379"
	}
}

func main() {
	flag.Parse()
	opts := broker.NewOption(broker.Leader, *endpoint, *savepath, *persistent, []string{etcdurl} )
	broker.NewBroker(opts).Run()
}
