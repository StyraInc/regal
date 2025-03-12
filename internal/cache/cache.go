package cache

import (
	"sync"

	"github.com/open-policy-agent/opa/v1/ast"
)

type baseCache struct {
	root *baseCacheElem
	rwm  sync.RWMutex
}

type baseCacheElem struct {
	value    ast.Value
	children map[ast.Value]*baseCacheElem
}

func NewBaseCache() *baseCache {
	return &baseCache{
		root: newBaseCacheElem(),
		rwm:  sync.RWMutex{},
	}
}

func (c *baseCache) Get(ref ast.Ref) ast.Value {
	c.rwm.RLock()
	defer c.rwm.RUnlock()

	node := c.root

	for i := range ref {
		node = node.children[ref[i].Value]
		if node == nil {
			return nil
		} else if node.value != nil {
			if len(ref) == 1 && ast.IsScalar(node.value) {
				// If the node is a scalar, return the value directly
				// and avoid an allocation when calling Find.
				return node.value
			}

			result, err := node.value.Find(ref[i+1:])
			if err != nil {
				return nil
			}

			return result
		}
	}

	return nil
}

func (c *baseCache) Put(ref ast.Ref, value ast.Value) {
	c.rwm.Lock()
	node := c.root

	for i := range ref {
		if child, ok := node.children[ref[i].Value]; ok {
			node = child
		} else {
			child := newBaseCacheElem()
			node.children[ref[i].Value] = child
			node = child
		}
	}

	node.set(value)
	c.rwm.Unlock()
}

func newBaseCacheElem() *baseCacheElem {
	return &baseCacheElem{
		children: map[ast.Value]*baseCacheElem{},
	}
}

func (e *baseCacheElem) set(value ast.Value) {
	e.value = value
	e.children = map[ast.Value]*baseCacheElem{}
}
