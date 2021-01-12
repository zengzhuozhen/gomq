package service

import (
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"
	"net"
	"time"
)

type MemberReceiver struct {
	MemberConnPool   map[string]net.Conn // Member连接
	MemberSyncMap    map[string]uint64   // Member最新同步量
	MemberQuitSignal chan string         // member remote address
	BroadcastChan    common.MsgUnitChan
	Queue            *common.Queue

	LP uint64 // 低水位
	HP uint64 // 高水位
}

func NewMemberReceiver(queue *common.Queue) *MemberReceiver {
	return &MemberReceiver{
		MemberConnPool:   make(map[string]net.Conn, 1000),
		MemberSyncMap:    make(map[string]uint64, 1000),
		MemberQuitSignal: make(chan string),
		BroadcastChan:    make(common.MsgUnitChan),
		Queue:            queue,
	}
}

func (m *MemberReceiver) Broadcast() error {
	for {
		select {
		case msg := <-m.BroadcastChan:
			messagePacket := msg.Pack()
			log.Debugf("Leader准备广播消息:{Topic:'%s'} {Body:'%s'}", msg.Topic, msg.Data.Body)
			for address, conn := range m.MemberConnPool {
				m.send(address, conn, messagePacket)
			}
		case address := <-m.MemberQuitSignal:
			log.Debugf("Member{socket:'%s'}连接关闭")
			delete(m.MemberConnPool, address)
		}
	}
}

func (m *MemberReceiver) send(address string, conn net.Conn, messagePacket []byte) {
	if m.isNewOne(address) {
		// 广播到每个member之前，先看看member是不是新来的，是的话先同步一下此前所有记录
		for topic, dataAssemble := range m.Queue.Local {
			for _, msg := range dataAssemble {
				if topic != msg.Topic {
					log.Debugf("主题和数据内容不符合")
				}
				if _, err := conn.Write(msg.Pack()); err != nil {
					log.Errorf(err.Error())
				}
			}
		}
	} else {
		if _, err := conn.Write(messagePacket); err != nil {
			log.Errorf(err.Error())
		}
	}
}



func (m *MemberReceiver) RegisterSyncAndResponse(conn net.Conn, packet *protocolPacket.SyncReqPacket) {
	m.MemberConnPool[conn.RemoteAddr().String()] = conn
	syncAckPacket := protocolPacket.NewSyncAckPacket(packet.PacketIdentifier)
	err := syncAckPacket.Write(conn)
	if err != nil {
		log.Errorf("返回syncAck失败", err)
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
	if len(m.MemberSyncMap) == 1 {
		return true
	}
	for a, i := range m.MemberSyncMap {
		if i <= offset && a != addr {
			return false
		}
	}
	return true
}
