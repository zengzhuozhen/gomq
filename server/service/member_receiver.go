package service

import (
	"fmt"
	protocolPacket "gomq/protocol/packet"
	"net"
)

type MemberReceiver struct {
	MemberConnPool map[string]net.Conn // Member连接
	MemberSyncMap  map[string]uint64   // Member最新同步量

	LowWaterPlane  uint64 // 低水位
	HighWaterPlane uint64 // 高水位
}

func NewMemberReceiver() *MemberReceiver {
	return &MemberReceiver{
		MemberConnPool: make(map[string]net.Conn, 1000),
		MemberSyncMap:  make(map[string]uint64, 1000),
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

func (m *MemberReceiver) SyncOffset(conn net.Conn, packet *protocolPacket.SyncOffsetPacket) {
	m.MemberSyncMap[conn.RemoteAddr().String()] = packet.Offset
}
