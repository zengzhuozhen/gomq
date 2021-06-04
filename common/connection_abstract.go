package common

type ConnectionAbstract struct {
	ConsumerConnAbs *ConsumerConnAbs
	// IsOldOne 判断是否为重启的旧客户端
	IsOldOne     bool
	WillFlag     bool
	WillQos      uint8
	WillRetain   bool
	WillTopic    string
	WillMessage  string
	UserNameFlag bool
	PasswordFlag bool
	KeepAlive    uint16
}


type ConsumerConnAbs struct {
	// Topic 按顺序存储的主题，暂时无法用TopPosMap替换，因为ChanAssemble需要依赖其顺序性
	Topic []string
	// TopPosMap topic和position的映射
	TopPosMap map[string]int64
}

func (connection *ConnectionAbstract) IsEmptyTopic() bool {
	return len(connection.ConsumerConnAbs.TopPosMap) == 0
}

// UpdatePosition update the topic position to `to` ,if `to` is negative,just++
func (connection *ConnectionAbstract) UpdatePosition(topic string, to int64) {
	if to >= 0 {
		connection.ConsumerConnAbs.TopPosMap[topic] = to
	} else {
		connection.ConsumerConnAbs.TopPosMap[topic]++
	}
}
