package xpath

import (
	"fmt"
	"strings"

	"github.com/midbel/angle/xml"
)

func Find(query string, doc xml.Node, ctx *Context) (Sequence, error) {
	p := NewParser(strings.NewReader(query))
	expr, err := p.Parse()
	if err != nil {
		return nil, err
	}
	it := newNodeItem(doc)
	return eval(expr, createFocus(it, 1, 1), ctx)
}

func eval(expr Expr, curr Focus, ctx *Context) (Sequence, error) {
	switch e := expr.(type) {
	case path:
		return evalPath(e, curr, ctx)
	case root:
		return evalRoot(e, curr, ctx)
	case step:
		return evalStep(e, curr, ctx)
	case nameTest:
		return evalName(e, curr, ctx)
	case predicate:
		return evalPredicate(e, curr, ctx)
	case typeTest:
		return evalType(e, curr, ctx)
	case binary:
		return evalBinary(e, curr, ctx)
	case unary:
		return evalUnary(e, curr, ctx)
	case number:
		return evalNumber(e, curr, ctx)
	case literal:
		return evalLiteral(e, curr, ctx)
	default:
		return nil, fmt.Errorf("unsupported expression")
	}
	return nil, nil
}

func evalPath(expr path, curr Focus, ctx *Context) (Sequence, error) {
	seq := Singleton(curr.Item)
	for i := range expr {
		var res Sequence
		for j := range seq {
			if seq[j].Atomic() {
				continue
			}
			sub, err := eval(expr[i], curr.Sub(seq[j], seq.Len(), j+1), ctx)
			if err != nil {
				return nil, err
			}
			res = append(res, sub...)
		}
		seq = res
	}
	return seq, nil
}

func evalRoot(_ root, curr Focus, ctx *Context) (Sequence, error) {
	node, err := currentNode(curr.Item)
	if err != nil {
		return nil, err
	}
	n := traverseRoot(node)
	if n == nil {
		return nil, nil
	}
	return Singleton(n), nil
}

func evalStep(expr step, curr Focus, ctx *Context) (Sequence, error) {
	switch expr.axis {
	default:
		return nil, fmt.Errorf("unsupported axis")
	case SelfAxis:
		return eval(expr.Expr, curr.Sub(curr.Item, 1, 1), ctx)
	case AttributeAxis:
		node, err := currentNode(curr.Item)
		if err != nil {
			return nil, err
		}
		var (
			seq Sequence
			attrs = traverseAttributes(node)
		)
		for i, a := range attrs {
			it := newNodeItem(a)
			res, err := eval(expr.Expr, curr.Sub(it, len(attrs), i+1), ctx)
			if err != nil {
				return nil, err
			}
			seq = append(seq, res...)
		}
		return seq, nil
	case ParentAxis:
		n, err := currentNode(curr.Item)
		if err != nil {
			return nil, err
		}
		if p := n.Parent(); p != nil {
			it := newNodeItem(p)
			return eval(expr.Expr, curr.Sub(it, 1, 1), ctx)
		}
	case ChildAxis:
		node, err := currentNode(curr.Item)
		if err != nil {
			return nil, err
		}
		var (
			nodes = traverseChild(node)
			seq   Sequence
		)
		for i := range nodes {
			it := newNodeItem(nodes[i])
			res, err := eval(expr.Expr, curr.Sub(it, len(nodes), i+1), ctx)
			if err != nil {
				return nil, err
			}
			seq = append(seq, res...)
		}
		return seq, nil
	case DescendantAxis, DescendantOrSelfAxis:
		node, err := currentNode(curr.Item)
		if err != nil {
			return nil, err
		}
		var (
			seq      Sequence
			traverse = traverseDeep
		)
		if expr.axis == DescendantOrSelfAxis {
			traverse = traverseDeepAndSelf
		}
		nodes := traverse(node, 0)
		for i := range nodes {
			it := newNodeItem(nodes[i])
			res, err := eval(expr.Expr, curr.Sub(it, len(nodes), i+1), ctx)
			if err != nil {
				return nil, err
			}
			seq = append(seq, res...)
		}
		return seq, nil
	}
	return nil, nil
}

func evalName(expr nameTest, curr Focus, ctx *Context) (Sequence, error) {
	node, err := currentNode(curr.Item)
	if err != nil {
		return nil, err
	}
	n, ok := getNameFromNode(node)
	if !ok {
		return nil, nil
	}
	if expr.Prefix != "" && expr.Prefix != "*" {
		if expr.Prefix != n.Prefix {
			return nil, nil
		}
	}
	if expr.Name != "" && expr.Name != "*" {
		if expr.Name != n.Local {
			return nil, nil
		}
	}
	return Singleton(node), nil
}

func evalType(expr typeTest, curr Focus, ctx *Context) (Sequence, error) {
	var nt xml.NodeType
	switch expr.Kind {
	case NodeTypeNode:
		return Singleton(curr.Item), nil
	case NodeTypeElem:
		nt = xml.ElementNode
	case NodeTypeAttr:
		nt = xml.AttrNode
	case NodeTypeText:
		nt = xml.TextNode
	case NodeTypeComment:
		nt = xml.CommentNode
	case NodeTypePI:
		nt = xml.PINode
	default:
		return nil, nil
	}
	node, err := currentNode(curr.Item)
	if err != nil {
		return nil, err
	}
	if t := node.Type(); t != nt {
		return nil, nil
	}
	return Singleton(node), nil
}

func evalPredicate(expr predicate, curr Focus, ctx *Context) (Sequence, error) {
	seq, err := eval(expr.Expr, curr, ctx)
	if err != nil {
		return nil, err
	}
	var res Sequence
	for i := range seq {
		sub, err := eval(expr.Check, curr.Sub(seq[i], seq.Len(), i+1), ctx)
		if err != nil {
			return nil, err
		}
		res = append(res, sub...)
	}
	return res, nil
}

func evalBinary(expr binary, curr Focus, ctx *Context) (Sequence, error) {
	return nil, nil
}

func evalUnary(expr unary, curr Focus, ctx *Context) (Sequence, error) {
	return nil, nil
}

func evalNumber(expr number, curr Focus, ctx *Context) (Sequence, error) {
	return nil, nil
}

func evalLiteral(expr literal, curr Focus, ctx *Context) (Sequence, error) {
	return nil, nil
}

func currentNode(item Item) (xml.Node, error) {
	if item.Atomic() {
		return nil, fmt.Errorf("node item expected")
	}
	return item.Node(), nil
}

func traverseAttributes(node xml.Node) []xml.Attribute {
	el, ok := node.(*xml.Element)
	if !ok {
		return nil
	}
	return el.Attributes
}

func traverseRoot(node xml.Node) xml.Node {
	root := node
	for {
		tmp := root.Parent()
		if tmp == nil {
			break
		}
		root = tmp
	}
	return root
}

func traverseDeepAndSelf(node xml.Node, depth int) []xml.Node {
	nodes := []xml.Node{node}
	others := traverseDeep(node, depth)
	return append(nodes, others...)
}

func traverseDeep(node xml.Node, depth int) []xml.Node {
	var nodes []xml.Node

	var walk func(xml.Node)
	walk = func(n xml.Node) {
		for _, child := range traverseChild(n) {
			nodes = append(nodes, child)
			walk(child)
		}
	}

	walk(node)
	return nodes
}

func traverseChild(node xml.Node) []xml.Node {
	switch n := node.(type) {
	case *xml.Document:
		return n.Children
	case *xml.Element:
		return n.Children
	default:
		return nil
	}
}

func getNameFromNode(node xml.Node) (xml.Name, bool) {
	switch n := node.(type) {
	case *xml.Element:
		return n.Name, true
	case *xml.PI:
		return n.Name, true
	default:
		return xml.Name{}, false
	}
}
