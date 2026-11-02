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
	return eval(expr, createContext(doc, 1, 1))
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
	case typeTest:
		return evalType(e, ctx)
	default:
		return nil, fmt.Errorf("unsupported expression")
	}
	return nil, nil
}

func evalPath(expr path, ctx Context) (Sequence, error) {
	seq := Singleton(ctx.Node)
	for i := range expr {
		var res Sequence
		for j := range seq {
			if seq[j].Atomic() {
				continue
			}
			n := seq[j].Node()

			sub, err := eval(expr[i], ctx.Sub(n, seq.Len(), j+1))
			if err != nil {
				return nil, err
			}
			res = append(res, sub...)
		}
		seq = res
	}
	return seq, nil
}

func evalRoot(_ root, ctx Context) (Sequence, error) {
	n := traverseRoot(ctx.Node)
	if n == nil {
		return nil, nil
	}
	return Singleton(n), nil
}

func evalStep(expr step, ctx Context) (Sequence, error) {
	switch expr.axis {
	default:
		return nil, fmt.Errorf("unsupported axis")
	case SelfAxis:
		return eval(expr.Expr, ctx.Sub(ctx.Node, 1, 1))
	case ParentAxis:
		p := ctx.Node.Parent()
		if p != nil {
			return eval(expr.Expr, ctx.Sub(p, 1, 1))
		}
	case ChildAxis:
		var (
			nodes = traverseChild(ctx.Node)
			seq   Sequence
		)
		for i := range nodes {
			res, err := eval(expr.Expr, ctx.Sub(nodes[i], len(nodes), i+1))
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

func evalType(expr typeTest, ctx Context) (Sequence, error) {
	var nt xml.NodeType
	switch expr.Kind {
	case NodeTypeNode:
		return Singleton(ctx.Node), nil
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
	if t := ctx.Node.Type(); t != nt {
		return nil, nil
	}
	return Singleton(ctx.Node), nil
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
