package store

import (
	"github.com/zengzhuozhen/gomq/common"
	"sync"
)

func NewMemStore() Store {
	return &memStore{
		locker: new(sync.RWMutex),
		isOpen: false,
		data:   make(map[string][]common.MessageUnit),
	}
}

type memStore struct {
	locker *sync.RWMutex
	isOpen bool
	data   map[string][]common.MessageUnit
}

func (m *memStore) Open(string) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if m.isOpen == false {
		m.isOpen = true
	}
}

func (m *memStore) Append(item common.MessageUnit) {
	if m.isOpen == false {
		panic("member store is not open")
	}
	m.locker.Lock()
	defer m.locker.Unlock()
	m.data[item.Topic] = append(m.data[item.Topic], item)

}

func (m *memStore) Reset(string) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.data = make(map[string][]common.MessageUnit)
}

func (m *memStore) ReadAll(topic string) []common.MessageUnit {
	return m.data[topic]
}

func (m *memStore) Close() {
	m.locker.RLock()
	defer m.locker.RUnlock()
	m.isOpen = false
}

func (m *memStore) Cap(topic string) int {
	return len(m.data[topic])
}


func (m *memStore)GetAllTopics () (topics []string){
	for topic , _ := range m.data{
		topics = append(topics, topic)
	}
	return
}