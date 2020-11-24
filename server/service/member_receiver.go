package service

import (
	"fmt"
	"gomq/common"
	protocolPacket "gomq/protocol/packet"
	"net"
	"time"
)

type MemberReceiver struct {
	MemberConnPool   map[string]net.Conn // Member连接
	MemberSyncMap    map[string]uint64   // Member最新同步量
	MemberQuitSignal chan string         // member remote address
	BroadcastChan    common.MsgUnitChan

	LP uint64 // 低水位
	HP uint64 // 高水位
}

func NewMemberReceiver() *MemberReceiver {
	return &MemberReceiver{
		MemberConnPool:   make(map[string]net.Conn, 1000),
		MemberSyncMap:    make(map[string]uint64, 1000),
		MemberQuitSignal: make(chan string),
		BroadcastChan:    make(common.MsgUnitChan),
	}
}

func (m *MemberReceiver) Broadcast() error {
	for {
		select {
		case msg := <-m.BroadcastChan:
			messagePacket := msg.Pack()
			fmt.Printf("Leader准备广播消息:{Topic:'%s'} {Body:'%s'}", msg.Topic, msg.Data.Body)
			for address, conn := range m.MemberConnPool {
				// 广播到每个member之前，先看看member是不是新来的，是的话先同步一下此前所有记录
				if m.isNewOne(address) {
					// todo 同步旧数据
				}
				if _, err := conn.Write(messagePacket); err != nil {
					fmt.Println(err)
				}
			}
		case address := <-m.MemberQuitSignal:
			fmt.Printf("Member{socket:'%s'}连接关闭")
			delete(m.MemberConnPool, address)
		}
	}
}

func (m *MemberReceiver) RegisterSyncAndResponse(conn net.Conn, packet *protocolPacket.SyncReqPacket) {
	m.MemberConnPool[conn.RemoteAddr().String()] = conn
	syncAckPacket := protocolPacket.NewSyncAckPacket(packet.PacketIdentifier)
	err := syncAckPacket.Write(conn)
	if err != nil {
		fmt.Println("返回syncAck失败", err)
	}
}

func (m *MemberReceiver) UpdateSyncOffset(conn net.Conn, packet *protocolPacket.SyncOffsetPacket) {
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	m.MemberSyncMap[conn.RemoteAddr().String()] = packet.Offset
	// 每次更新member的offset,环视一周看看其他member的水位是不是比当前member的高，是则拉高低水位线
	if m.isLowestOne(conn.RemoteAddr().String(), packet.Offset) {
		m.LP = packet.Offset
	}
}

func (m *MemberReceiver) isNewOne(address string) bool {
	return m.MemberSyncMap[address] == 0
}

func (m *MemberReceiver) isLowestOne(addr string, offset uint64) bool {
	for a, i := range m.MemberSyncMap {
		if i <= offset && a != addr {
			return false
		}
	}
	return true
}