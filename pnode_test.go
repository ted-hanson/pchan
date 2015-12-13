package pchan

import (
	"math/rand"
	"testing"
)

// Generic new node
func TestNewPNode(t *testing.T) {
	i := rand.New(rand.NewSource(12345)).Int()

	pnode := NewPNode(i)

	if pnode.Priority != i {
		t.Fatalf("pnode priority should have been %d but is %d", i, pnode.Priority)
	}

	if pnode.Count != 0 {
		t.Fatalf("pnode Count should start at 0 but is %d", pnode.Count)
	}

	testObj := "test"
	go func() {
		pnode.Values <- testObj
	}()

	select {
	case x := <-pnode.Values:
		if x != testObj {
			t.Fatalf("pnode Values channel was not initialized to an empty channel and had: %#v", x)
		}
	}

	return
}

func TestPNodeCount(t *testing.T) {
	i := rand.New(rand.NewSource(12345)).Int()

	pnode := NewPNode(i)

	for i := 0; i < 100; i += 1 {
		_ = pnode.RequestAddToValues()
		if int(pnode.GetCount()) != i+1 {
			t.Fatalf("pnode incremented incorrectly! expected count to be %d but was %d", i+1, pnode.GetCount())
		}
	}

	for i := 0; i < 100; i += 1 {
		_ = pnode.RequestPopFromValues()
		if int(pnode.GetCount()) != 99-i {
			t.Fatalf("pnode decremented incorrectly! expected count to be %d but got %d", 99-i, pnode.GetCount())
		}
	}

	_ = pnode.RequestPopFromValues()
	if int(pnode.GetCount()) != 0 {
		t.Fatalf("pnode decremented incorrectly! expected count to be 0 but got %d", pnode.GetCount())
	}

	return
}
