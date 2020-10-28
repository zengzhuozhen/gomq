package main

import (
	"gomq/server/broker"
	_ "net/http/pprof"
)




func main() {
	opts := broker.NewOption(broker.Leader,"127.0.0.1:9001",[]string{"127.0.0.1:2379"})
	broker.NewBroker(opts).Run()
}
