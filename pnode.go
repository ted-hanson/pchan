package pchan

import (
	"sync"
	"sync/atomic"
)

// ============  PNODE =============
// A PNode has 3 fields:
//  - Priority: priority of Values chan
//  - Values: chan of generic objects
//  - Count: valid count of Queuing objects
type PNode struct {
	Priority int
	Values   chan interface{}
	Count    uint64

	// Lock to make sure Count is >= # objects in `Values`
	sync.RWMutex
}

// Generic new node
func NewPNode(p int) (newNode *PNode) {
	return &PNode{
		Priority: p,
		Values:   make(chan interface{}),
		Count:    0,
	}
}

func (n *PNode) decrement() {
	atomic.AddUint64(&n.Count, ^uint64(0))
}

func (n *PNode) increment() {
	atomic.AddUint64(&n.Count, 1)
}

func (n *PNode) GetCount() uint64 {
	return atomic.LoadUint64(&n.Count)
}

func (n *PNode) RequestPopFromValues() chan interface{} {
	n.Lock()
	if n.GetCount() > 0 {
		n.decrement()
		n.Unlock()
		return n.Values
	} else {
		n.Unlock()
		return nil
	}
}

func (n *PNode) RequestAddToValues() chan interface{} {
	n.Lock()
	n.increment()
	n.Unlock()
	return n.Values
}
