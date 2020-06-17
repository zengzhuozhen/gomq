package consumer

import (
	"sync"
)

type Pool struct {
	ConnUids []string
	State    map[string]bool
	Position map[string][]int64
	Topic    map[string][]string
	mu       sync.Mutex
}

func NewPool() *Pool {
	return &Pool{
		ConnUids: []string{},
		State:    make(map[string]bool, 1024),
		Position: make(map[string][]int64, 1024),
		Topic:    make(map[string][]string, 1024),
		mu:       sync.Mutex{},
	}
}

func (p *Pool) ForeachActiveConn() []string {
	connUids := make([]string, 0)
	for _, i := range p.ConnUids {
		if p.State[i] == true && len(p.Topic[i]) != 0 {
			connUids = append(connUids, i)
		}
	}
	return connUids
}

func (p *Pool) Add(connUid string, topics []string) {
	p.ConnUids = append(p.ConnUids, connUid)
	p.State[connUid] = true
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
