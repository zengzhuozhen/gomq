package main

import (
	"flag"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/server/broker"
	_ "net/http/pprof"
)

var endpoint = flag.String("endpoint", ":9000", "mq服务运行地址")
var dirname = flag.String("dirname", "/var/log/tempmq/", "mq数据保存文件夹路径")
var persistent = flag.Bool("persistent", true, "是否开启持久化")
var etcdurl string

func init() {
	if common.IsRunningInDocker() {
		etcdurl = "http://etcd:2379"
	} else {
		etcdurl = "127.0.0.1:2379"
	}
}


// todo 零拷贝 copyBuffer , impl writeTo and ReadFrom
func main() {
	flag.Parse()
	opts := broker.NewOption(broker.Leader, *endpoint, *dirname, []string{etcdurl})
	broker.NewBroker(opts).Run()
}
