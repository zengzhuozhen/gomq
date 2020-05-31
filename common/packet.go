package common

type Flag int

const (
	C  Flag = iota + 1 // consumer
	P                  // producer
	S                  // server
	CH                 // consumer heart
)

type Packet struct {
	Flag    Flag    `json:"flag"`
	Message Message `json:"message"`
}

func NewPacket(flag Flag, message Message) Packet {
	return Packet{
		Flag:    flag,
		Message: message,
	}
}
