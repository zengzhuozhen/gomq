package consumer

import (
	"sync"
)

type Pool struct {
	ConnUids []string
	Position map[string]int64
	State    map[string]bool
	Topic    map[string]string
	mu       sync.Mutex
}

func NewPool() *Pool {
	return &Pool{
		ConnUids: []string{},
		Position: make(map[string]int64, 1024),
		State:    make(map[string]bool, 1024),
		Topic:    make(map[string]string, 1024),
		mu:       sync.Mutex{},
	}
}

func (p *Pool) ForeachActiveConn() []string {
	connUids := make([]string,0)
	for _, i := range p.ConnUids {
		if p.State[i] == true && p.Topic[i] != "" {
			connUids = append(connUids,i)
		}
	}
	return connUids
}

func (p *Pool) Add(connUid, topic string,position int64) {
	p.ConnUids = append(p.ConnUids, connUid)
	p.Position[connUid] = position
	p.State[connUid] = true
	p.Topic[connUid] = topic
}

func (p *Pool) UpdatePosition(uid string){
	p.mu.Lock()
	p.Position[uid]++
	p.mu.Unlock()
}