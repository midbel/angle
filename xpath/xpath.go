package xpath

import (
	"fmt"
	"strconv"

	"github.com/midbel/angle/xml"
)

const (
	nsPrefixFunc = "fn"
	nsUriFunc    = "http://www.w3.org/2005/xpath-functions"
)

type Env[T any] map[string]T

func (e Env[T]) Resolve(ident string) (T, error) {
	t, ok := e[ident]
	if !ok {
		return t, fmt.Errorf("%s: undefined identifier", ident)
	}
	return t, nil
}

func (e Env[T]) Define(ident string, value T) {
	e[ident] = value
}

type Context struct {
	namespaces           Env[string]
	vars                 Env[any]
	defaultElemNamespace string
	defaultFuncNamespace string
}

func NewContext() *Context {
	return &Context{
		namespaces:           make(Env[string]),
		vars:                 make(Env[any]),
		defaultFuncNamespace: nsUriFunc,
	}
}

func (c *Context) ResolveVar(ident string) (any, error) {
	return c.vars.Resolve(ident)
}

func (c *Context) RegisterVar(ident string, value any) {
	c.vars.Define(ident, value)
}

func (c *Context) RegisterNS(ident, value string) {
	c.namespaces.Define(ident, value)
}

func (c *Context) SetDefaultElemNamespace(uri string) {
	c.defaultElemNamespace = uri
}

func (c *Context) SetDefaultFuncNamespace(uri string) {
	c.defaultFuncNamespace = uri
}

func (c *Context) ResolveElementNamespace(prefix string) (string, error) {
	if prefix == "" {
		return c.defaultElemNamespace, nil
	}
	uri, err := c.namespaces.Resolve(prefix)
	if err != nil {
		return "", fmt.Errorf("%s: unregistered uri prefix", prefix)
	}
	return uri, nil
}

func (c *Context) ResolveFunctionNamespace(prefix string) (string, error) {
	if prefix == "" {
		return c.defaultFuncNamespace, nil
	}
	if prefix == nsPrefixFunc {
		return nsUriFunc, nil
	}
	uri, err := c.namespaces.Resolve(prefix)
	if err != nil {
		return "", fmt.Errorf("%s: unregistered uri prefix", prefix)
	}
	return uri, nil
}

type qname struct {
	Uri  string
	Name string
}

func createQName(name, prefix string) qname {
	var uri string
	switch prefix {
	case nsPrefixFunc:
		uri = nsUriFunc
	default:
	}
	return qname{
		Uri:  uri,
		Name: name,
	}
}

func (q qname) prefix() string {
	switch q.Uri {
	case nsUriFunc:
		return nsPrefixFunc
	default:
		return ""
	}
}

func (q qname) expandedName() string {
	prefix := q.prefix()
	if prefix == "" {
		return q.Name
	}
	return fmt.Sprintf("%s:%s", prefix, q.Name)
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

func (s Sequence) First() Item {
	if s.Empty() {
		return nil
	}
	return s[0]
}

func (s Sequence) Last() Item {
	if s.Empty() {
		return nil
	}
	return s[s.Len()-1]
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
	return i.node
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

func (atomicItem) Atomic() bool {
	return true
}

type untyped string

func atomize(value any) (any, error) {
	switch n := value.(type) {
	case xml.Node:
		c := n.Content()
		return untyped(c), nil
	default:
		return value, nil
	}
}

func stringify(value any) (string, error) {
	v, err := atomize(value)
	if err != nil {
		return "", err
	}
	if u, ok := v.(untyped); ok {
		return string(u), nil
	}
	switch v := v.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case atomicItem:
		return stringify(v.Value())
	case nodeItem:
		return stringify(v.Node())
	default:
		return "", err
	}
}
