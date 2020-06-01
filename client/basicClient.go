package client

import (
	"log"
	"net"
	"strconv"
	"time"
)

type BasicClient struct {
	Protocol string
	Host     string
	Port     int
	Timeout  int // timeout sec
}

func (b *BasicClient) Connect() net.Conn {
	conn, err := net.DialTimeout(b.Protocol,b.Host+ ":" + strconv.Itoa(b.Port), time.Duration(b.Timeout) * time.Second)
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}
	return conn
}


