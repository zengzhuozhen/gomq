package store

import "sync"

func NewMemStore() Store{
	return new(memStore)
}

type memStore struct {
	locker sync.RWMutex
	isOpen  bool
	data  []PersistentUnit
}



func (m *memStore) Open() {
	m.locker.Lock()
	defer m.locker.Unlock()
	if m.isOpen == true{
		panic("already open the member store")
	}else {
		m.data = make([]PersistentUnit,0)
		m.isOpen = true
	}
}

func (m *memStore) Append(item PersistentUnit) {
	if m.isOpen == false{
		panic("member store is not open")
	}
	m.locker.Lock()
	defer m.locker.Unlock()
	m.data = append(m.data,item)
}

func (m *memStore) SnapShot() {
	return
}

func (m *memStore) Reset() {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.data = make([]PersistentUnit,0)
}

func (m *memStore) Load() {
	return
}

func (m *memStore) Close() {
	m.locker.RLock()
	defer m.locker.RUnlock()
	if m.isOpen == false{
		panic("mem")
	}
}

func (m *memStore) Cap() int {
	return len(m.data)
}