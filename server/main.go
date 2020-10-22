package main

import (
	"gomq/server/broker"
	_ "net/http/pprof"
)




func main() {
	broker.NewBroker(broker.Leader).Run()
}
