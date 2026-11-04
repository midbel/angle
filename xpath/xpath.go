package xpath

import (
	"fmt"

	"github.com/midbel/angle/xml"
)

type Env[T any] map[string]T

func (e Env[T]) Resolve(ident string) (T, error) {
	var t T
	return t, nil
}

func (e Env[T]) Define(ident string, value T) {
	e[ident] = value
}

type Context struct {
	namespaces Env[string]
	vars       Env[string]
}

func NewContext() *Context {
	return &Context{
		namespaces: make(Env[string]),
		vars:       make(Env[string]),
	}
}

type Focus struct {
	Item
	Index int
	Size  int
}

func createFocus(item Item, size, ix int) Focus {
	return Focus{
		Item:  item,
		Index: ix,
		Size:  size,
	}
}

func (c Focus) Sub(item Item, size, ix int) Focus {
	x := Focus{
		Item:  item,
		Size:  size,
		Index: ix,
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
	if it, ok := value.(Item); ok {
		return Sequence{it}
	}
	var it Item
	switch v := value.(type) {
	case nodeItem:
	case atomicItem:
	case xml.Node:
		it = newNodeItem(v)
	default:
		it = newAtomicItem(value)
	}
	return Sequence{it}
}

func (s Sequence) Empty() bool {
	return s.Len() == 0
}

func (s Sequence) Singleton() bool {
	return s.Len() == 1
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
	return i.Node
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

func atomize(node xml.Node) (string, error) {
	switch n := node.(type) {
	case *xml.Element:
		return n.Content(), nil
	case xml.Attribute:
		return n.Value, nil
	default:
		return "", fmt.Errorf("node can not be atomized")
	}
}
