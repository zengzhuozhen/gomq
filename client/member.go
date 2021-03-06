package client

import (
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/protocol"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"time"
)

type Member struct {
	opts           *Option
	client         *client
	SyncOffset     uint64
	PersistentChan chan common.MessageUnit

	recentMsgBuffer map[string]struct{}
}

func NewMember(opts *Option) *Member {
	client := newClient(opts)
	return &Member{
		client:          client,
		opts:            opts,
		SyncOffset:      0,
		PersistentChan:  make(chan common.MessageUnit),
		recentMsgBuffer: make(map[string]struct{}, 1000),
	}
}

func (m *Member) SendSync() error {
	err := m.client.connect()
	if err != nil {
		panic("连接Leader失败")
	}
	syncPacket := protocolPacket.NewSyncReqPacket(1)
	err = syncPacket.Write(m.client.conn)
	if err != nil {
		fmt.Println("Member请求同步失败")
	}
	fmt.Println("发送syncreq")

	go m.SendSyncOffset(3 * time.Second)
	go m.ReadPacket()
	return nil
}

func (m *Member) SendSyncOffset(duration time.Duration) {
	tickTimer := time.NewTicker(duration)
	for {
		select {
		case <-tickTimer.C:
			pingReqPack := protocolPacket.NewSyncOffsetPacket(m.SyncOffset, m.client.getAvailableIdentity())
			pingReqPack.Write(m.client.conn)
		}
	}
}

func (m *Member) ReadPacket() {
	for {
		// 读取数据包
		var fh protocolPacket.FixedHeader
		if err := fh.Read(m.client.conn); err != nil {
			fmt.Errorf("读取包头失败%+v", err)
			return
		}
		switch fh.TypeAndReserved {
		case protocol.SYNCACK:
			var syncackPacket protocolPacket.SyncAckPacket
			err := syncackPacket.Read(m.client.conn, fh)
			if err != nil {
				fmt.Println("Leader返回SyncAck失败")
				m.client.conn.Close()
			}
			fmt.Println("接收到Leader返回的SyncAck消息")
		default:
			// 同步普通消息
			messByte := make([]byte, 1024 * 20) // 分配足够大空间保证包不会被截断
			n, _ := m.client.conn.Read(messByte)
			head := fh.Pack()
			data := append(head.Bytes(), messByte[:n]...)
			message := new(common.MessageUnit)
			message = message.UnPack(data)
			if !m.checkMsgExist(message.Data.GetId()) {
				m.PersistentChan <- *message
				m.SyncOffset++
			}
			// drop the message
		}
	}
}

func (m *Member) checkMsgExist(MsgId string) bool {
	if _, ok := m.recentMsgBuffer[MsgId]; ok == true {
		return true
	}
	// todo Bloom Filter to validate key exists

	return false
}
