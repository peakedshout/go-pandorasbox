package tmap

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func Load[T1 any, T2 any](m Map, key T1) (load T2, ok bool) {
	l, b := m.Load(key)
	if !b {
		return load, b
	}
	return l.(T2), b
}

// Store sets the value for a key.
func Store[T1 any, T2 any](m Map, key T1, value T2) {
	m.Store(key, value)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func LoadOrStore[T1 any, T2 any](m Map, key T1, value T2) (actual T2, loaded bool) {
	store, l := m.LoadOrStore(key, value)
	return store.(T2), l
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func LoadAndDelete[T1 any, T2 any](m Map, key T1) (value T2, loaded bool) {
	v, l := m.LoadAndDelete(key)
	if !l {
		return value, l
	}
	return v.(T2), l
}

// Delete deletes the value for a key.
func Delete[T1 any, T2 any](m Map, key T1) {
	m.Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently (including by f), Range may reflect any
// mapping for that key from any point during the Range call. Range does not
// block other methods on the receiver; even f itself may call any method on m.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func Range[T1 any, T2 any](m Map, fn func(key T1, value T2) (shouldContinue bool)) {
	m.Range(func(key, value any) (shouldContinue bool) {
		return fn(key.(T1), value.(T2))
	})
}
