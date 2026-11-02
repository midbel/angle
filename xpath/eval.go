package xpath

import (
	"fmt"
	"io"

	"github.com/midbel/angle/xml"
)

func Find(r io.Reader, doc xml.Node, ctx *Context) (Sequence, error) {
	p := NewParser(r)
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
	case name:
		return evalName(e, curr, ctx)
	case typeTest:
		return evalType(e, curr, ctx)
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
	}
	return nil, nil
}

func evalName(expr name, curr Focus, ctx *Context) (Sequence, error) {
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
	return Singleton(curr.Node), nil
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

func currentNode(item Item) (xml.Node, error) {
	if item.Atomic() {
		return nil, fmt.Errorf("node item expected")
	}
	return item.Node(), nil
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
