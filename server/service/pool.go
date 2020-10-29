package service

import (
	"sync"
)

type Pool struct {
	ConnUids []string
	State    *sync.Map
	// position 和 topic 总是相对应的
	Position map[string][]int64		// Position[ConnId]{Topic_A_position,Topic_B_position,Topic_C_position}
	Topic    map[string][]string    // Topic[ConnId]{Topic_A,Topic_B,Topic_C}
	mu       sync.Mutex
}

func NewPool() *Pool {
	return &Pool{
		ConnUids: []string{},
		State:    new(sync.Map),
		Position: make(map[string][]int64, 1024),
		Topic:    make(map[string][]string, 1024),
		mu:       sync.Mutex{},
	}
}

func (p *Pool) ForeachActiveConn() []string {
	connUids := make([]string,0)
	for _, i := range p.ConnUids {
		isActive, ok := p.State.Load(i)
		if !ok {
			panic("active not exist in state pool")
		}
		if isActive == true && len(p.Topic[i]) != 0 {
			connUids = append(connUids, i)
		}
	}
	return connUids
}

func (p *Pool) Add(connUid string, topics []string) {
	p.ConnUids = append(p.ConnUids, connUid)
	p.State.Store(connUid, true)
	p.Position[connUid] = make([]int64, len(topics))
	p.Topic[connUid] = topics
}

func (p *Pool) UpdatePosition(uid, topic string) {
	p.mu.Lock()
	for k, v := range p.Topic[uid] {
		if topic == v {
			p.Position[uid][k]++
			break
		}
	}
	p.mu.Unlock()
}
