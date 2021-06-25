# GoMQ

Golang实现的消息队列

## 目录

- [GoMQ](#GoMQ)
    - [运行](#运行)
        - [依赖](#依赖)
        - [占用端口](#占用端口)
        - [Docker](#Docker)
        - [命令行工具](#命令行工具)
        - [HTTP](#HTTP)
    - [设计](#设计)
        - [关键概念](#关键概念)
        - [整体架构](#整体架构)
    - [整体](#整体)
        - [benchmark](#benchmark)
    - [TODO](#TODO)

## 运行

### 依赖

|工具|仓库|版本|作用|
|----|----|----|----|
|ETCD|https://github.com/etcd-io/etcd|v3.3.25+incompatible|注册运行的MQ实例

### 占用端口

|端口|作用|
|----|----|
|9000|消息队列服务监听的TCP端口|
|8000|消息队列服务监听的HTTP端口|
|2379|ETCD默认占用端口|
|2380|ETCD默认占用端口|
|4001|ETCD默认占用端口|
|7001|ETCD默认占用端口|

### Docker

推荐使用Docker运行

构建容器：make dockerPrepare

运行容器：docker-compose up -d

### 命令行工具

可以用命令行工具进行操作

USAGE:

```
gomqctl [global options] command [command options] [arguments...]
```

COMMANDS:

- list show messages
- pub publish message
- sub subscribe message
- version get current version
- help, h Shows a list of commands or help for one command

GLOBAL OPTIONS:

- --connect value connect the server broker
- --qos value set the message qos (default: 0)
- --retain value set the message retain (default: 0)
- --topic value the topic you care
- --help, -h show help (default: false)

发布消息：

```
docker exec -it gomq gomqctl --topic A --connect 127.0.0.1:9000 pub hello wrold everyone
```

订阅消息：

```
docker exec -it gomq gomqctl --topic A --connect 127.0.0.1:9000 sub 
```

### HTTP

    略

### 作为客户端使用

```
import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zengzhuozhen/gomq/client"
	"github.com/zengzhuozhen/gomq/common"
)

func main() {
	opt := &client.Option{
		Protocol: "tcp",
		Address:  "127.0.0.1:9000",
		KeepAlive:  30,
	}
	producer := client.NewProducer(opt)
	msg := common.MessageUnit{
		Topic: "A",
		QoS:   0,
		Data:  common.Message{
			Id:     uuid.New().String(),
			MsgKey: "key",
			Body:   "hello world",
		},
	}
	producer.Publish(msg,0,1)

	customer := client.NewConsumer(opt)

	for i := range customer.Subscribe([]string{"A"},0){
		fmt.Printf("收到主题:%s的消息:%s",i.Topic,i.Data.Body)
	}
}
```

#### GoMod 依赖问题
由于依赖ETCD，仓库可能会出现一下问题：
> go: github.com/coreos/bbolt@v1.3.4: parsing go.mod:
module declares its path as: go.etcd.io/bbolt
but was required as: github.com/coreos/bbolt

解决方法：
> replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

## 设计

### 关键概念

|术语|说明|
   |---|---|
|QoS（服务质量等级）| 分为最多一次，最少一次，精确一次|
|retain（保留消息）| 决定消息是否持久化|
|consumer| 消息的消费者，可以订阅多个主题|
|producer|消息的生产者，可以发布消息到指定主题|
|leader | 服务集群的leader,负责MQ主要功能，并同步数据给member|
|member | 服务集群的member,只负责拷贝leader的数据|

### 整体架构

    略

## 整体

### benchmark

    略

## TODO

- [ ] Broker断线后保存未处理信息(Qos>0)
- [ ] 主从切换高可用
- [x] 日志压缩算法
- [ ] 日志分片
- [ ] 持久化消息存储规划
- [ ] 客户端消费偏移持久化