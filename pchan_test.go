package pchan

import (
	"sync/atomic"
	"testing"
)

var tests int
var cases int

func init() {
	tests = 100000
	cases = 5
}

func TestReceiveSendLimitedPriorities(t *testing.T) {
	pchan := NewPChan()

	calcSum := 0
	// chan to return when one goroutine finishes
	done := make(chan int)
	for x := 0; x < tests; x += 1 {
		y := x % cases
		calcSum += y
		go func() {
			go func() {
				done <- 1
			}()
			pchan.Send(y, y)
		}()
	}

	// block until all goroutines are done
	completed := make(chan int)
	go func() {
		for x := 0; x < tests; x += 1 {
			<-done
		}
		// all goroutines done!
		completed <- 1
	}()
	<-completed

	// Validate that the number of priorities not being used is correct (before pulling)
	if !(pchan.priorityHeap.Len() == cases-1 || pchan.priorityHeap.Len() == tests-1) {
		t.Fatalf("# of priorities in priorityHeap == %d when it should be %d or %d", pchan.priorityHeap.Len(), tests-1, cases-1)
	}

	actualSum := 0
	for x := 0; x < tests; x += 1 {
		actualSum += pchan.Receive().(int)
	}

	if int(pchan.pendingCount) != 0 {
		t.Fatalf("# of pendingCount should be %d but was %d", 0, pchan.pendingCount)
	}
	if pchan.priorityHeap.Len() != 0 {
		t.Fatalf("# of priorities in priorityHeap == %d when it should be %d or %d", pchan.priorityHeap.Len(), tests-1, cases-1)
	}

	if actualSum != calcSum {
		t.Fatalf("Sum of received numbers sum([1..%d]) = %d did not equal the sent amount: %d", tests, actualSum, calcSum)
	}
}

func TestReceiveSend(t *testing.T) {
	pchan := NewPChan()
	actualSum := 0
	calcSum := tests * (tests - 1) / 2
	for x := 0; x < tests; x += 1 {
		y := x
		go func() {
			pchan.Send(y, y)
		}()
	}
	for x := 0; x < tests; x += 1 {
		actualSum += pchan.Receive().(int)
	}
	if actualSum != calcSum {
		t.Fatalf("Sum of received numbers sum([1..%d]) = %d did not equal the sent amount: %d", tests, actualSum, calcSum)
	}
}

func TestSendReceive(t *testing.T) {
	pchan := NewPChan()

	var actualSum uint64 = 0
	calcSum := uint64(tests * (tests - 1) / 2)

	// chan to count goroutines
	done := make(chan int)
	// chan when goroutines are done
	completed := make(chan int)
	// spin up a `tests` amount of goroutines
	for x := 0; x < tests; x += 1 {
		go func() {
			atomic.AddUint64(&actualSum, pchan.Receive().(uint64))
			// 1 goroutine done!
			done <- 1
		}()
	}

	go func() {
		for x := 0; x < tests; x += 1 {
			<-done
		}
		// all goroutines done!
		completed <- 1
	}()

	for x := 0; x < tests; x += 1 {
		pchan.Send(x, uint64(x))
	}

	<-completed
	if actualSum != calcSum {
		t.Fatalf("Sum of received numbers sum([1..%d]) = %d did not equal the sent amount: %d", tests, actualSum, calcSum)
	}
}

func TestSendReceiveLimitedPriorities(t *testing.T) {
	pchan := NewPChan()

	var actualSum uint64 = 0
	var calcSum uint64 = 0

	// chan to count goroutines
	done := make(chan int)
	// chan when goroutines are done
	completed := make(chan int)
	// spin up a `tests` amount of goroutines
	for x := 0; x < tests; x += 1 {
		go func() {
			atomic.AddUint64(&actualSum, pchan.Receive().(uint64))
			done <- 1
		}()
	}

	for {
		// make sure pendingCount eventually gets to the number of tests
		if int(pchan.pendingCount) != tests {
			break
		}
	}

	for x := 0; x < tests; x += 1 {
		y := x % cases
		calcSum += uint64(y)
		pchan.Send(y, uint64(y))
	}

	go func() {
		for x := 0; x < tests; x += 1 {
			<-done
		}
		// all goroutines done!
		completed <- 1
	}()
	<-completed

	if pchan.priorityHeap.Len() != 0 {
		t.Fatalf("Priority Heap should be length 0 is but has %d elements!\n%#v", pchan.priorityHeap.Len(), pchan.priorityHeap)
	}

	if actualSum != calcSum {
		t.Fatalf("Sum of received numbers sum([1..%d]) = %d did not equal the sent amount: %d", tests, actualSum, calcSum)
	}
}
