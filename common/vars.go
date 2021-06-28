package common

import (
	"flag"
)

var Endpoint = flag.String("Endpoint", "127.0.0.1:9000", "mq服务运行地址")
var Dirname = flag.String("Dirname", "/var/log/gomq/", "mq数据保存文件夹路径")
var EtcdUrl string

func init() {
	if IsRunningInDocker() {
		EtcdUrl = "http://etcd:2379"
	} else {
		EtcdUrl = "127.0.0.1:2379"
	}
}
