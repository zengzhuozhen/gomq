package store

import "github.com/zengzhuozhen/gomq/common"


type Store interface {
	// Open 打开存储器
	Open(topic string)
	// Append 追加数据
	Append(item common.MessageUnit)
	// Reset 重置
	Reset(topic string)
	// ReadAll 读取数据
	ReadAll(topic string) []common.MessageUnit
	// Close 关闭存储器
	Close()
	// Cap 容量
	Cap(topic string) int
	// GetAllTopics 获取所有主题
	GetAllTopics()[]string
}


