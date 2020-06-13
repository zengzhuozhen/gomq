package producer

import (
	"bytes"
	"fmt"
	"gomq/client"
	"gomq/common"
	"gomq/protocol/utils"
	"testing"
)


func TestProducer_Publish_Topic_A(t *testing.T) {
	producer := 	NewProducer("tcp","127.0.0.1",9000,10)
	mess := common.Message{
		MsgId: 1,
		MsgKey: "test",
		Body:   "hello world A ",
	}
	producer.Publish("A",mess,2)
}

func TestProducer_Publish_Topic_B(t *testing.T) {
	producer := 	NewProducer("tcp","127.0.0.1",9000,10)
	mess := common.Message{
		MsgId: 2,
		MsgKey: "test",
		Body:   "hello world B",
	}
	producer.Publish("B",mess,1)
}

func TestProducer_Publish_Topic_C(t *testing.T) {
	producer := 	NewProducer("tcp","127.0.0.1",9000,10)
	mess := common.Message{
		MsgId: 3,
		MsgKey: "test",
		Body:   "hello world C",
	}
	producer.Publish("C",mess,2)
}

func TestName(t *testing.T) {

		bc := client.BasicClient{
			Protocol:     "tcp",
			Host:         "127.0.0.1",
			Port:         9000,
			Timeout:      110,
			IdentityPool: nil,
		}
		net,_ := bc.Connect()
		var body bytes.Buffer
		body.Write(utils.EncodeString("test"))
		var header bytes.Buffer
		header.WriteByte(1)
		header.Write([]byte{2})
		header.Write(body.Bytes())
		_, _ = net.Write(header.Bytes())

		fmt.Println(net)
}

