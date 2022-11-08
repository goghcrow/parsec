package example

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/goghcrow/parsec"
)

// 绑定缓存的生命周期与 state 绑定在一起
var cacheG = &sync.Map{} // state => cache

func getCache(s parsec.State) cache {
	k := ptrOf(s)
	store, loaded := cacheG.LoadOrStore(k, cache{})
	if !loaded {
		runtime.SetFinalizer(s, func(_ interface{}) {
			_, ok := cacheG.Load(k)
			if ok {
				cacheG.Delete(k)
			}
		})
	}
	return store.(cache)
}

type cacheItem struct {
	result interface{} //result or error
	err    error
	rest   parsec.Loc
}

type cache map[string]*cacheItem

func (c cache) key(p parsec.Parser, s parsec.State) string {
	return fmt.Sprintf("%p_%d", p, s.Save().Pos)
}

func (c cache) get(p parsec.Parser, s parsec.State) *cacheItem {
	if m, ok := cacheG.Load(ptrOf(s)); ok {
		return m.(cache)[c.key(p, s)]
	}
	return nil
}

func (c cache) put(p parsec.Parser, s parsec.State, v interface{}, err error) {
	item := &cacheItem{
		result: v,
		err:    err,
		rest:   s.Save(),
	}
	k := ptrOf(s)
	m, ok := cacheG.Load(k)
	if !ok {
		m = map[string]*cacheItem{}
		cacheG.Store(k, m)
	}
	m.(cache)[c.key(p, s)] = item
}

func ptrOf(ptr interface{}) uintptr { return reflect.ValueOf(ptr).Pointer() }
