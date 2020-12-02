package store

import (
	"gomq/common"
	"sync"
)

func NewMemStore() Store{
	return &memStore{
		locker: new(sync.RWMutex),
		isOpen: false,
		data:   make([]common.MessageUnit,0),
	}
}

type memStore struct {
	locker *sync.RWMutex
	isOpen  bool
	data  []common.MessageUnit
}



func (m *memStore) Open() {
	m.locker.Lock()
	defer m.locker.Unlock()
	if m.isOpen == true{
		return
	}else {
		m.isOpen = true
	}
}

func (m *memStore) Append(item common.MessageUnit) {
	if m.isOpen == false{
		panic("member store is not open")
	}
	m.locker.Lock()
	defer m.locker.Unlock()
	m.data = append(m.data,item)
}


func (m *memStore) Reset() {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.data = make([]common.MessageUnit,0)
}

func (m *memStore) Load() {
	return
}

func (m *memStore) Close() {
	m.locker.RLock()
	defer m.locker.RUnlock()
	m.isOpen = false
}

func (m *memStore) Cap() int {
	return len(m.data)
}