package client

import (
	"fmt"
	"gomq/common"
	"gomq/protocol"
	protocolPacket "gomq/protocol/packet"
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
	client := NewClient(opts)
	return &Member{
		client:          client,
		opts:            opts,
		SyncOffset:      0,
		PersistentChan:  make(chan common.MessageUnit),
		recentMsgBuffer: make(map[string]struct{}, 1000),
	}
}

func (m *Member) SendSync() error {
	err := m.client.Connect()
	fmt.Println(m.client.conn.LocalAddr().String())
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
			fmt.Println("发送同步消息量数据包")
			pingReqPack := protocolPacket.NewSyncOffsetPacket(m.SyncOffset, m.client.GetAvailableIdentity())
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
			messByte := make([]byte, 4096) // todo fix:这里可能由于粘包导致超出slice长度
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
