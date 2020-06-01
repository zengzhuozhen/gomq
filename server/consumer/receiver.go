package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"gomq/common"
	"log"
	"net"
)



type Receiver struct {
	chanAssemble map[string]common.MsgChan
	pool         *Pool
}

func NewConsumerReceiver(chanAssemble map[string]common.MsgChan, pool *Pool) *Receiver {
	return &Receiver{chanAssemble: chanAssemble, pool: pool}
}

func (r *Receiver)HandleConn(netPacket common.Packet,conn net.Conn){
	connUid :=conn.RemoteAddr().String()
  	r.pool.Add(connUid,netPacket.Topic,netPacket.Position)
	r.chanAssemble[connUid] = make(common.MsgChan,1000)
}

func (r *Receiver)HandleQuit(connUid string){
	r.pool.State[connUid] = false
	close(r.chanAssemble[connUid])
	delete(r.chanAssemble, connUid)
}

func (r *Receiver) Consume(ctx context.Context,conn net.Conn) {
	connUid := conn.RemoteAddr().String()
	for {
		select {
		case msg := <-r.chanAssemble[connUid]:
			// 防止多个管道同时竞争所有消息的问题,采用客户端连接池进行逻辑隔离解决
			netPacket := common.Packet{
				Flag:    common.S,
				Message: msg,
				Topic:   r.pool.Topic[connUid],
			}
			if data, err := json.Marshal(netPacket); err == nil {
				fmt.Printf("准备发送给客户端消费者: %s %+v", conn.RemoteAddr(), netPacket)
				n, err := conn.Write(data)
				fmt.Println(n)
				if err !=nil{
					fmt.Println(err)
				}
			} else {
				log.Fatalln(err.Error())
			}
		case <-ctx.Done():
			fmt.Println("消费者连接已关闭，退出消费循环")
			r.HandleQuit(connUid)
			return
		}
	}
}


