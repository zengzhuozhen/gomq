package store

import "gomq/common"

// 持久化单元
type PersistentUnit struct {
	topic string
	data  common.Message
}

func NewPersistentUnit(topic string,data common.Message) PersistentUnit {
	return PersistentUnit{
		topic: topic,
		data:  data,
	}
}

type Store interface {
	// 打开存储器
	Open()
	// 追加数据
	Append(item PersistentUnit)
	// 快照
	SnapShot()
	// 重置
	Reset()
	// 加载到内存
	Load()
	// 关闭存储器
	Close()
	// 容量
	Cap() int
}
