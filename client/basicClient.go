package client

import (
	"log"
	"net"
	"strconv"
	"time"
)

type BasicClient struct {
	protocol string
	host string
	port int
	timeout int		// timeout sec
}

func (b *BasicClient) Connect() net.Conn {
	conn, err := net.DialTimeout(b.protocol,b.host + ":" + strconv.Itoa(b.port), time.Duration(b.timeout) * time.Second)
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}
	return conn
}


