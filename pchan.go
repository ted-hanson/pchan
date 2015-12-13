package pchan

import (
	"container/heap"
	"sync"
	"sync/atomic"
)

// ============  PCHAN =============
// A PChan has 4 fields:
//  - curnNode: Next Node to serve
//  - priorityHeap: heap of priorities
//  - pendingCount: Count of Sent elements that haven't been received
//  - pendingValues: Channel to hold Sent elements
//  - priorityToNode: Map of priority to Node
type PChan struct {
	curNode        *PNode
	priorityHeap   *IntHeap
	pendingCount   uint64
	pendingValues  chan interface{}
	priorityToNode map[int]*PNode
	sync.Mutex
}

// Generate new PChan
func NewPChan() (pchan *PChan) {
	pchan = &PChan{
		curNode:        nil,
		priorityHeap:   &IntHeap{},
		pendingCount:   0,
		pendingValues:  make(chan interface{}),
		priorityToNode: make(map[int]*PNode),
	}
	heap.Init(pchan.priorityHeap)
	return
}

func (t *PChan) incrementPending() {
	atomic.AddUint64(&t.pendingCount, 1)
}

func (t *PChan) decrementPending() {
	atomic.AddUint64(&t.pendingCount, ^uint64(0))
}

func (t *PChan) getPending() uint64 {
	return atomic.LoadUint64(&t.pendingCount)
}

// !!! PRECONDITIONS !!!
// AT LEAST has a read lock on pchan.curNode
// pchan.curNode is not nil
// pchan.curNode is not going to receive anymore
func (pchan *PChan) popNode() (next *PNode) {
	// remove current empty PNode
	delete(pchan.priorityToNode, pchan.curNode.Priority)
	// pop off priorityHeap to get next priority
	if pchan.priorityHeap.Len() > 0 {
		return pchan.priorityToNode[heap.Pop(pchan.priorityHeap).(int)]
	}
	return nil
}

func (pchan *PChan) Receive() (nextVal interface{}) {
	pchan.Lock()
	for {
		// No Sent values yet: put in pending
		if pchan.curNode == nil {
			pchan.incrementPending()
			pchan.Unlock()
			return <-pchan.pendingValues
		} else {
			// Grab Sent values from PNodes
			select {
			case nextVal = <-pchan.curNode.RequestPopFromValues():
				pchan.Unlock()
				return
			default:
				oldNode := pchan.curNode
				oldNode.RLock()
				// Replace node if Count is empty
				if oldNode.GetCount() == 0 {
					pchan.curNode = pchan.popNode()
				}
				oldNode.RUnlock()
			}
		}
	}
}

func (pchan *PChan) addNode(p int) (newNode *PNode) {
	newNode = NewPNode(p)
	pchan.priorityToNode[p] = newNode
	if pchan.curNode == nil {
		pchan.curNode = newNode
	} else {
		// push lower priority node into heap and update curNode
		if newNode.Priority > pchan.curNode.Priority {
			heap.Push(pchan.priorityHeap, pchan.curNode.Priority)
			pchan.curNode = newNode
		} else {
			heap.Push(pchan.priorityHeap, newNode.Priority)
		}
	}
	return
}

func (pchan *PChan) Send(p int, v interface{}) {
	pchan.Lock()
	if pchan.getPending() > 0 {
		pchan.decrementPending()
		pchan.Unlock()
		pchan.pendingValues <- v
		return
	}
	node, ok := pchan.priorityToNode[p]
	if !ok {
		node = pchan.addNode(p)
	}
	values := node.RequestAddToValues()
	pchan.Unlock()
	// Actually send v over the priority's channel
	values <- v
}
