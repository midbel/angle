package xpath

import (
	"github.com/midbel/angle/xml"
)

type Context struct {
	xml.Node
	Index int
	Size  int

	namespaces map[string]string
	env        map[string]string
}

func createContext(node xml.Node, size, ix int) Context {
	return Context{
		Node:       node,
		Index:      ix,
		Size:       size,
		namespaces: make(map[string]string),
		env:        make(map[string]string),
	}
}

func (c Context) Sub(node xml.Node, size, ix int) Context {
	x := Context{
		Node:       node,
		Size:       size,
		Index:      ix,
		namespaces: c.namespaces,
		env:        c.env,
	}
	return x
}

type Item interface {
	Node() xml.Node
	Value() any
	True() bool
	Atomic() bool
}

type Sequence []Item

func Singleton(value any) Sequence {
	var it Item
	switch v := value.(type) {
	case xml.Node:
		it = newNodeItem(v)
	default:
		it = newAtomicItem(value)
	}
	return Sequence{it}
}

func (s Sequence) Len() int {
	return len(s)
}

type nodeItem struct {
	node xml.Node
}

func newNodeItem(node xml.Node) Item {
	return nodeItem{
		node: node,
	}
}

func (i nodeItem) Node() xml.Node {
	return i.node
}

func (i nodeItem) Value() any {
	return nil
}

func (i nodeItem) True() bool {
	return false
}

func (nodeItem) Atomic() bool {
	return false
}

type atomicItem struct {
	value any
}

func newAtomicItem(value any) Item {
	return atomicItem{
		value: value,
	}
}

func (atomicItem) Node() xml.Node {
	return nil
}

func (a atomicItem) Value() any {
	return a.value
}

func (a atomicItem) True() bool {
	return false
}

func (atomicItem) Atomic() bool {
	return true
}
