package xpath

import (
	"github.com/midbel/angle/xml"
)

type Context struct {
	xml.Node
	Index int
	Size  int
}

type Item interface {
	Node() xml.Node
	Value() any
	True() bool
	Atomic() bool
}

type Sequence []Item

type nodeItem struct {
	node xml.Node
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
