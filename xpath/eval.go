package xpath

import (
	"fmt"
	"io"

	"github.com/midbel/angle/xml"
)

func Find(r io.Reader, doc xml.Node) (Sequence, error) {
	p := NewParser(r)
	expr, err := p.Parse()
	if err != nil {
		return nil, err
	}
	ctx := Context{
		Node:  doc,
		Index: 1,
		Size:  1,
	}
	return eval(expr, ctx)
}

func eval(expr Expr, ctx Context) (Sequence, error) {
	switch e := expr.(type) {
	case path:
		return evalPath(e, ctx)
	case root:
		return evalRoot(e, ctx)
	case step:
		return evalStep(e, ctx)
	case name:
		return evalName(e, ctx)
	default:
		return nil, fmt.Errorf("unsupported expression")
	}
	return nil, nil
}

func evalPath(expr path, ctx Context) (Sequence, error) {
	return nil, nil
}

func evalRoot(expr root, ctx Context) (Sequence, error) {
	return nil, nil
}

func evalStep(expr step, ctx Context) (Sequence, error) {
	switch expr.axis {
	default:
		return nil, fmt.Errorf("unsupported axis")
	case ChildAxis:
		var (
			nodes = traverseChild(ctx.Node)
			sub   Context
			seq   Sequence
		)
		sub.Size = len(nodes)
		for i := range nodes {
			sub.Index = i + 1
			sub.Node = nodes[i]

			res, err := eval(expr.Expr, sub)
			if err != nil {
				return nil, err
			}
			seq = append(seq, res...)
		}
		return seq, nil
	}
	return nil, nil
}

func evalName(expr name, ctx Context) (Sequence, error) {
	n, ok := getNameFromNode(ctx.Node)
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
	return Singleton(ctx.Node), nil
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
