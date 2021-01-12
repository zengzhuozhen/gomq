package store

import "github.com/zengzhuozhen/gomq/common"


type Store interface {
	// 打开存储器
	Open(topic string)
	// 追加数据
	Append(item common.MessageUnit)
	// 重置
	Reset(topic string)
	// 读取数据
	ReadAll(topic string) []common.MessageUnit
	// 关闭存储器
	Close()
	// 容量
	Cap(topic string) int
}


