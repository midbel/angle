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
