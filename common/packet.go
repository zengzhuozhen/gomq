package common

type Flag int

const (
	C  Flag = iota + 1 // consumer
	P                  // producer
	S                  // server
	CH                 // consumer heart
)

type Packet struct {
	Flag     Flag    `json:"flag"`
	Message  Message `json:"message"`
	Topic    string  `json:"topic"`
	Position int64   `json:"position"` // only consume need set
}

