package visit

import "github.com/zengzhuozhen/gomq/protocol/packet"

type PacketVisitor struct {
	Packet packet.ControlPacket
}

func (p *PacketVisitor) Visit(fn packet.VisitorFunc) error {
	return fn(p.Packet)
}
