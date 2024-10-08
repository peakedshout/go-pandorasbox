package tmap

import (
	"sync"
)

type TSyncMap[T1, T2 any] struct {
	m sync.Map
}

func NewTSyncMap[T1, T2 any]() TSyncMap[T1, T2] {
	t := TSyncMap[T1, T2]{m: sync.Map{}}
	return t
}

func NewTSyncMapPtr[T1, T2 any]() *TSyncMap[T1, T2] {
	m := NewTSyncMap[T1, T2]()
	return &m
}

func (tsm *TSyncMap[T1, T2]) Load(key T1) (value T2, ok bool) {
	load, ok := tsm.m.Load(key)
	if !ok {
		return value, false
	}
	return load.(T2), true
}

func (tsm *TSyncMap[T1, T2]) Store(key T1, value T2) {
	tsm.m.Store(key, value)
}
func (tsm *TSyncMap[T1, T2]) LoadOrStore(key T1, value T2) (actual T2, loaded bool) {
	v, loaded := tsm.m.LoadOrStore(key, value)
	if v != nil {
		actual = v.(T2)
	}
	return actual, loaded
}
func (tsm *TSyncMap[T1, T2]) LoadAndDelete(key T1) (value T2, loaded bool) {
	v, loaded := tsm.m.LoadAndDelete(key)
	if loaded {
		return v.(T2), true
	}
	return value, false
}
func (tsm *TSyncMap[T1, T2]) Delete(key T1) {
	tsm.m.Delete(key)
}
func (tsm *TSyncMap[T1, T2]) Range(f func(key T1, value T2) bool) {
	tsm.m.Range(func(key, value any) bool {
		return f(key.(T1), value.(T2))
	})
}
