package store

import "gomq/common"



type Store interface {
	// 打开存储器
	Open()
	// 追加数据
	Append(item common.MessageUnit)
	// 重置
	Reset()
	// 加载到内存
	Load()
	// 关闭存储器
	Close()
	// 容量
	Cap() int
}
