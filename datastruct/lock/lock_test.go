package lock

import (
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMakeInitializesTable(t *testing.T) {
	locks := Make(8)
	if locks == nil || len(locks.table) != 8 {
		t.Fatalf("Make(8) = %#v, want 8-entry table", locks)
	}
	for i, mu := range locks.table {
		if mu == nil {
			t.Fatalf("table[%d] = nil, want initialized *sync.RWMutex", i)
		}
	}
}

func TestFnv32KnownVectors(t *testing.T) {
	tests := []struct {
		in   string
		want uint32
	}{
		{"", 0x811c9dc5}, // offset basis; loop never runs
		{"a", 0x050c5d7e},
		{"foobar", 0x31f0b262},
		{"godis", 0xebed08dd},
	}
	for _, tt := range tests {
		if got := fnv32(tt.in); got != tt.want {
			t.Fatalf("fnv32(%q) = %#x, want %#x", tt.in, got, tt.want)
		}
	}
}

func TestFnv32IsDeterministicAndDistinct(t *testing.T) {
	if fnv32("redis") == fnv32("godis") {
		t.Fatal("fnv32 should distinguish different inputs")
	}
	if fnv32("redis") != fnv32("redis") {
		t.Fatal("fnv32 should be deterministic")
	}
}

func TestSpreadStaysWithinTable(t *testing.T) {
	locks := Make(16)
	for _, key := range []string{"a", "b", "foobar", "godis", "redis", "long-key-name"} {
		idx := locks.spread(fnv32(key))
		if idx >= uint32(len(locks.table)) {
			t.Fatalf("spread(%q) = %d, must be < %d", key, idx, len(locks.table))
		}
	}
}

func TestLockProvidesMutualExclusion(t *testing.T) {
	locks := Make(16)
	key := "shared"
	var counter int
	var inFlight int32
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locks.Lock(key)
			cur := atomic.AddInt32(&inFlight, 1)
			if cur > 1 {
				t.Errorf("concurrent holders = %d, want <= 1", cur)
			}
			counter++
			atomic.AddInt32(&inFlight, -1)
			locks.UnLock(key)
		}()
	}
	wg.Wait()
	if counter != 100 {
		t.Fatalf("counter = %d, want 100", counter)
	}
}

func TestRLockAllowsTwoConcurrentReaders(t *testing.T) {
	locks := Make(16)
	key := "ro"
	locks.RLock(key)
	defer locks.RUnLock(key)

	done := make(chan struct{})
	go func() {
		locks.RLock(key)
		defer locks.RUnLock(key)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second RLock blocked; RLock should allow concurrent readers")
	}
}

func TestWriteLockBlocksReader(t *testing.T) {
	locks := Make(16)
	key := "rw"
	locks.Lock(key)

	acquired := make(chan struct{})
	go func() {
		locks.RLock(key)
		defer locks.RUnLock(key)
		close(acquired)
	}()
	select {
	case <-acquired:
		t.Fatal("RLock acquired while write lock held")
	case <-time.After(50 * time.Millisecond):
	}

	locks.UnLock(key)
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("RLock did not acquire after write lock released")
	}
}

func TestToLockIndicesDedupsSameKey(t *testing.T) {
	locks := Make(16)
	indices := locks.toLockIndices([]string{"same", "same", "same"}, false)
	if len(indices) != 1 {
		t.Fatalf("expected 1 deduped index, got %v", indices)
	}
}

func TestToLockIndicesSortsAscending(t *testing.T) {
	locks := Make(16)
	indices := locks.toLockIndices([]string{"a", "b", "c", "d"}, false)
	if !sort.SliceIsSorted(indices, func(i, j int) bool { return indices[i] < indices[j] }) {
		t.Fatalf("indices not sorted ascending: %v", indices)
	}
}

func TestToLockIndicesSortsDescendingWhenReverse(t *testing.T) {
	locks := Make(16)
	indices := locks.toLockIndices([]string{"a", "b", "c", "d"}, true)
	if !sort.SliceIsSorted(indices, func(i, j int) bool { return indices[i] > indices[j] }) {
		t.Fatalf("indices not sorted descending: %v", indices)
	}
}

func TestLocksAndUnLocksBalanced(t *testing.T) {
	locks := Make(16)
	locks.Locks("a", "b", "c")
	// no panic / no deadlock; release in reverse order via UnLocks
	locks.UnLocks("a", "b", "c")

	// after release, a fresh exclusive lock must be acquirable
	locks.Lock("a")
	locks.UnLock("a")
}

func TestLocksDoesNotDeadlockOnOverlappingKeys(t *testing.T) {
	locks := Make(16)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			locks.Locks("a", "b", "c")
			locks.UnLocks("a", "b", "c")
		}()
		go func() {
			defer wg.Done()
			locks.Locks("c", "b", "a") // reversed arg order
			locks.UnLocks("c", "b", "a")
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Locks deadlocked on overlapping keys")
	}
}

func TestRLocksAndRUnLocksBalanced(t *testing.T) {
	locks := Make(16)
	locks.RLocks("x", "y")
	locks.RUnLocks("x", "y")

	// readers can re-enter after release
	locks.RLock("x")
	locks.RUnLock("x")
}

func TestRWLocksTreatsKeyInBothListsAsWrite(t *testing.T) {
	locks := Make(16)
	locks.RWLocks([]string{"k"}, []string{"k"})

	acquired := make(chan struct{})
	go func() {
		locks.RLock("k")
		defer locks.RUnLock("k")
		close(acquired)
	}()
	select {
	case <-acquired:
		t.Fatal("RLock acquired while RWLocks holds key as write")
	case <-time.After(50 * time.Millisecond):
	}

	locks.RWUnLocks([]string{"k"}, []string{"k"})
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("RLock did not acquire after RWUnLocks")
	}
}

func TestRWLocksAllowsConcurrentReaderForReadKeys(t *testing.T) {
	locks := Make(16)
	locks.RWLocks(nil, []string{"r"})
	defer locks.RWUnLocks(nil, []string{"r"})

	acquired := make(chan struct{})
	go func() {
		locks.RLock("r")
		defer locks.RUnLock("r")
		close(acquired)
	}()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("RLock blocked while only read lock held via RWLocks")
	}
}

func TestSpreadPanicsOnNilLocks(t *testing.T) {
	var nilLocks *Locks
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("spread on nil Locks should panic")
		}
	}()
	nilLocks.spread(0)
}
