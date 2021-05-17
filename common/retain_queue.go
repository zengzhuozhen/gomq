package common

type RetainQueue struct {
	ToPersistentChan chan MessageUnit
	readAllFunc      func(topic string) []MessageUnit
	resetFunc        func(topic string)
	capFunc          func(topic string) int
}

// NewRetainQueue return retain queue
func NewRetainQueue(readAllFunc func(topic string) []MessageUnit, resetFunc func(topic string), capFunc func(topic string) int) *RetainQueue {
	return &RetainQueue{
		ToPersistentChan: make(chan MessageUnit),
		readAllFunc:      readAllFunc,
		resetFunc:        resetFunc,
		capFunc:          capFunc,
	}
}

func (q *RetainQueue) Push(messageUnit MessageUnit) {
	q.ToPersistentChan <- messageUnit
}

func (q *RetainQueue) ReadAll(topic string) []MessageUnit {
	return q.readAllFunc(topic)
}

func (q *RetainQueue) Reset(topic string) {
	q.resetFunc(topic)
}

func (q *RetainQueue) Cap(topic string) int {
	return q.capFunc(topic)
}
