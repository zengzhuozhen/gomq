package service

import protocolPacket "github.com/zengzhuozhen/gomq/protocol/packet"

type ConnectionAbstract struct {
	// Position 按顺序存储的主题偏移，与Topic一一对应
	Position    []int64
	// Topic 按顺序存储的主题，与Position一一对应
	Topic       []string
	// IsOldOne 判断是否为重启的旧客户端
	IsOldOne    bool
	WillFlag    bool
	WillQos     int32
	WillRetain  bool
	WillTopic   string
	WillMessage protocolPacket.PublishPacket
}

func (connection *ConnectionAbstract) IsEmptyTopic() bool {
	return len(connection.Topic) == 0
}

// UpdatePosition update the topic position to `to` ,if `to` is negative,just++
func (connection *ConnectionAbstract)UpdatePosition(topic string,to int64){
	for k, v := range connection.Topic {
		if topic == v {
			if to > 0 {
				connection.Position[k] = to
			}else{
				connection.Position[k]++
			}
			break
		}
	}
}