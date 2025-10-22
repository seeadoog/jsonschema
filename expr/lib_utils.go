package expr

import (
	"context"
	"sync"
)

type cacheElement[K comparable, C any, V any] struct {
	k K
	c C
	v V
}

type InstanceCache[K comparable, C any, V any] struct {
	data        map[K]*cacheElement[K, C, V]
	newInstance func(ctx context.Context, k K, c C) (v V, err error)
	lock        sync.RWMutex
}

func (ic *InstanceCache[K, C, V]) Get(ctx context.Context, k K, c C) (v V, err error) {
	ic.lock.RLock()
	e := ic.data[k]
	ic.lock.RUnlock()
	if e != nil {
		return e.v, nil
	}
	ic.lock.Lock()
	defer ic.lock.Unlock()
	e = ic.data[k]
	if e != nil {
		return e.v, nil
	}
	v, err = ic.newInstance(ctx, k, c)
	if err != nil {
		return v, err
	}
	e = &cacheElement[K, C, V]{
		k: k,
		c: c,
		v: v,
	}
	ic.data[k] = e
	return v, nil
}
