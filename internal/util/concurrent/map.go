package concurrent

import (
	"maps"
	"sync"
)

type Map[K comparable, V any] struct {
	m    map[K]V
	murw sync.RWMutex
}

type ValueTransformer[V any] func(V) V

func MapOf[K comparable, V any](src map[K]V) *Map[K, V] {
	return &Map[K, V]{
		m:    src,
		murw: sync.RWMutex{},
	}
}

func (cm *Map[K, V]) Get(k K) (V, bool) {
	cm.murw.RLock()

	v, ok := cm.m[k]

	cm.murw.RUnlock()

	return v, ok
}

func (cm *Map[K, V]) GetUnchecked(k K) V {
	cm.murw.RLock()

	v := cm.m[k]

	cm.murw.RUnlock()

	return v
}

func (cm *Map[K, V]) Set(k K, v V) {
	cm.murw.Lock()

	cm.m[k] = v

	cm.murw.Unlock()
}

func (cm *Map[K, V]) Delete(k K) {
	cm.murw.Lock()

	delete(cm.m, k)

	cm.murw.Unlock()
}

func (cm *Map[K, V]) Keys() []K {
	cm.murw.RLock()

	keys := make([]K, 0, len(cm.m))

	for k := range cm.m {
		keys = append(keys, k)
	}

	cm.murw.RUnlock()

	return keys
}

func (cm *Map[K, V]) Values() []V {
	cm.murw.RLock()

	vs := make([]V, len(cm.m))
	i := 0

	for _, v := range cm.m {
		vs[i] = v
		i++
	}

	cm.murw.RUnlock()

	return vs
}

func (cm *Map[K, V]) Len() int {
	if cm == nil {
		return 0
	}

	cm.murw.RLock()

	l := len(cm.m)

	cm.murw.RUnlock()

	return l
}

func (cm *Map[K, V]) Clone() map[K]V {
	cm.murw.RLock()

	m := maps.Clone(cm.m)

	cm.murw.RUnlock()

	return m
}

func (cm *Map[K, V]) Clear() {
	cm.murw.Lock()

	clear(cm.m)

	cm.murw.Unlock()
}

func (cm *Map[K, V]) UpdateValue(key K, transformer ValueTransformer[V]) {
	cm.murw.Lock()

	var v V

	if vo, ok := cm.m[key]; ok {
		v = vo
	}

	cm.m[key] = transformer(v)

	cm.murw.Unlock()
}
