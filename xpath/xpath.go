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
	vars                 Env[string]
	defaultElemNamespace string
	defaultFuncNamespace string
}

func NewContext() *Context {
	return &Context{
		namespaces:           make(Env[string]),
		vars:                 make(Env[string]),
		defaultFuncNamespace: nsUriFunc,
	}
}

func (c *Context) RegisterVar(ident, value string) {
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
	return i.node
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

func convertTo(left, right Item) (any, any, error) {
	e, err := atomize(left.Value())
	if err != nil {
		return nil, nil, err
	}
	r, err := atomize(right.Value())
	if err != nil {
		return nil, nil, err
	}
	if _, ok := e.(untyped); ok {
		if _, ok := r.(untyped); ok {
			return string(e.(untyped)), string(r.(untyped)), nil
		}
		x, err := convertUntyped(e.(untyped), r)
		return x, r, err
	}
	if r, ok := r.(untyped); ok {
		x, err := convertUntyped(r, e)
		return e, x, err
	}
	switch e := e.(type) {
	case float64:
		x, ok := r.(float64)
		if ok {
			return e, x, nil
		}
	case int64:
		x, ok := r.(int64)
		if ok {
			return e, x, nil
		}
	case string:
		x, ok := r.(string)
		if ok {
			return e, x, nil
		}
	case bool:
		x, ok := r.(bool)
		if ok {
			return e, x, nil
		}
	default:
	}
	return nil, nil, ErrConvert
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

func convertUntyped(value untyped, target any) (any, error) {
	switch target.(type) {
	case float64:
		return toFloat(string(value))
	case int64:
		return toInt(string(value))
	case string:
		return toString(string(value))
	case bool:
		return toBool(string(value))
	default:
		return nil, ErrConvert
	}
}

func toFloat(value string) (float64, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		err = ErrConvert
	}
	return v, err
}

func toInt(value string) (int64, error) {
	v, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		err = ErrConvert
	}
	return v, err
}

func toString(value string) (string, error) {
	return value, nil
}

func toBool(value string) (bool, error) {
	b, err := strconv.ParseBool(value)
	if err == nil {
		err = ErrConvert
	}
	return b, err
}

func castToFloat(value any) (float64, error) {
	if u, ok := value.(untyped); ok {
		return castToFloat(string(u))
	}
	switch v := value.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, ErrConvert
	}
}
