package tmap

type TStdMap map[any]any

func MakeTStdMap(m map[any]any) TStdMap {
	if m == nil {
		return make(TStdMap)
	}
	return m
}

func (tm TStdMap) Load(key any) (any, bool) {
	a, ok := tm[key]
	return a, ok
}

func (tm TStdMap) Store(key, value any) {
	tm[key] = value
}

func (tm TStdMap) LoadOrStore(key, value any) (actual any, loaded bool) {
	load, b := tm.Load(key)
	if b {
		return load, b
	}
	tm.Store(key, value)
	return value, b
}

func (tm TStdMap) LoadAndDelete(key any) (value any, loaded bool) {
	load, b := tm.Load(key)
	tm.Delete(key)
	return load, b
}

func (tm TStdMap) Delete(a any) {
	delete(tm, a)
}

func (tm TStdMap) Range(f func(key any, value any) (shouldContinue bool)) {
	for k, v := range tm {
		if !f(k, v) {
			break
		}
	}
}
