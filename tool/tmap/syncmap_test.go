package tmap

import (
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"testing/quick"
)

type mapOp string

const (
	opLoad          = mapOp("Load")
	opStore         = mapOp("Store")
	opLoadOrStore   = mapOp("LoadOrStore")
	opLoadAndDelete = mapOp("LoadAndDelete")
	opDelete        = mapOp("Delete")
)

var mapOps = [...]mapOp{opLoad, opStore, opLoadOrStore, opLoadAndDelete, opDelete}

// mapCall is a quick.Generator for calls on mapInterface.
type mapCall struct {
	op   mapOp
	k, v any
}

func (c mapCall) apply(m Map) (any, bool) {
	switch c.op {
	case opLoad:
		return m.Load(c.k)
	case opStore:
		m.Store(c.k, c.v)
		return nil, false
	case opLoadOrStore:
		return m.LoadOrStore(c.k, c.v)
	case opLoadAndDelete:
		return m.LoadAndDelete(c.k)
	case opDelete:
		m.Delete(c.k)
		return nil, false
	default:
		panic("invalid mapOp")
	}
}

type mapResult struct {
	value any
	ok    bool
}

func randValue(r *rand.Rand) any {
	b := make([]byte, r.Intn(4))
	for i := range b {
		b[i] = 'a' + byte(rand.Intn(26))
	}
	return string(b)
}

func (mapCall) Generate(r *rand.Rand, size int) reflect.Value {
	c := mapCall{op: mapOps[rand.Intn(len(mapOps))], k: randValue(r)}
	switch c.op {
	case opStore, opLoadOrStore:
		c.v = randValue(r)
	}
	return reflect.ValueOf(c)
}

func applyCalls(m Map, calls []mapCall) (results []mapResult, final map[any]any) {
	for _, c := range calls {
		v, ok := c.apply(m)
		results = append(results, mapResult{v, ok})
	}

	final = make(map[any]any)
	m.Range(func(k, v any) bool {
		final[k] = v
		return true
	})

	return results, final
}

func applySyncMap(calls []mapCall) ([]mapResult, map[any]any) {
	return applyCalls(new(SyncMap[any, any]), calls)
}

func applyRWMutexMap(calls []mapCall) ([]mapResult, map[any]any) {
	return applyCalls(new(RWMutexMap), calls)
}

func applyDeepCopyMap(calls []mapCall) ([]mapResult, map[any]any) {
	return applyCalls(new(DeepCopyMap), calls)
}

func TestSyncMapMatchesRWMutex(t *testing.T) {
	if err := quick.CheckEqual(applySyncMap, applyRWMutexMap, nil); err != nil {
		t.Error(err)
	}
}

func TestSyncMapMatchesDeepCopy(t *testing.T) {
	if err := quick.CheckEqual(applySyncMap, applyDeepCopyMap, nil); err != nil {
		t.Error(err)
	}
}

func TestSyncMapConcurrentRange(t *testing.T) {
	const mapSize = 1 << 10

	//m := new(sync.Map)
	m := NewTSyncMap[int64, int64]()

	for n := int64(1); n <= mapSize; n++ {
		m.Store(n, int64(n))
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	defer func() {
		close(done)
		wg.Wait()
	}()
	for g := int64(runtime.GOMAXPROCS(0)); g > 0; g-- {
		r := rand.New(rand.NewSource(g))
		wg.Add(1)
		go func(g int64) {
			defer wg.Done()
			for i := int64(0); ; i++ {
				select {
				case <-done:
					return
				default:
				}
				for n := int64(1); n < mapSize; n++ {
					if r.Int63n(mapSize) == 0 {
						m.Store(n, n*i*g)
					} else {
						m.Load(n)
					}
				}
			}
		}(g)
	}

	iters := 1 << 10
	if testing.Short() {
		iters = 16
	}
	for n := iters; n > 0; n-- {
		seen := make(map[int64]bool, mapSize)

		m.Range(func(ki, vi int64) bool {
			k, v := ki, vi
			if v%k != 0 {
				t.Fatalf("while Storing multiples of %v, Range saw value %v", k, v)
			}
			if seen[k] {
				t.Fatalf("Range visited key %v twice", k)
			}
			seen[k] = true
			return true
		})

		if len(seen) != mapSize {
			t.Fatalf("Range visited %v elements of %v-element Map", len(seen), mapSize)
		}
	}
}

func TestSyncMapIssue40999(t *testing.T) {
	//var m sync.Map
	m := NewTSyncMap[*int, struct{}]()

	// Since the miss-counting in missLocked (via Delete)
	// compares the miss count with len(m.dirty),
	// add an initial entry to bias len(m.dirty) above the miss count.
	m.Store(nil, struct{}{})

	var finalized uint32

	// Set finalizers that count for collected keys. A non-zero count
	// indicates that keys have not been leaked.
	for atomic.LoadUint32(&finalized) == 0 {
		p := new(int)
		runtime.SetFinalizer(p, func(*int) {
			atomic.AddUint32(&finalized, 1)
		})
		m.Store(p, struct{}{})
		m.Delete(p)
		runtime.GC()
	}
}

func TestSyncMapRangeNestedCall(t *testing.T) { // Issue 46399
	//var m sync.Map
	m := NewTSyncMap[int, string]()

	for i, v := range [3]string{"hello", "world", "Go"} {
		m.Store(i, v)
	}
	m.Range(func(key int, value string) bool {
		m.Range(func(key int, value string) bool {
			// We should be able to load the key offered in the Range callback,
			// because there are no concurrent Delete involved in this tested map.
			if v, ok := m.Load(key); !ok || !reflect.DeepEqual(v, value) {
				t.Fatalf("Nested Range loads unexpected value, got %+v want %+v", v, value)
			}

			// We didn't keep 42 and a value into the map before, if somehow we loaded
			// a value from such a key, meaning there must be an internal bug regarding
			// nested range in the Map.
			if _, loaded := m.LoadOrStore(42, "dummy"); loaded {
				t.Fatalf("Nested Range loads unexpected value, want store a new value")
			}

			// Try to Store then LoadAndDelete the corresponding value with the key
			// 42 to the Map. In this case, the key 42 and associated value should be
			// removed from the Map. Therefore any future range won't observe key 42
			// as we checked in above.
			val := "sync.Map"
			m.Store(42, val)
			if v, loaded := m.LoadAndDelete(42); !loaded || !reflect.DeepEqual(v, val) {
				t.Fatalf("Nested Range loads unexpected value, got %v, want %v", v, val)
			}
			return true
		})

		// Remove key from Map on-the-fly.
		m.Delete(key)
		return true
	})

	// After a Range of Delete, all keys should be removed and any
	// further Range won't invoke the callback. Hence length remains 0.
	length := 0
	m.Range(func(key int, value string) bool {
		length++
		return true
	})

	if length != 0 {
		t.Fatalf("Unexpected sync.Map size, got %v want %v", length, 0)
	}
}
