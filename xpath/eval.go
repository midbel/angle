package xpath

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/midbel/angle/xml"
)

var (
	ErrComparable = errors.New("values are not comparable")
	ErrEval       = errors.New("expression can not be evaluated")
	ErrAtomic     = errors.New("atomic value")
	ErrConvert    = errors.New("value can not be converted")
)

func compareError(left, right any) error {
	return fmt.Errorf("%w: %v, %v", ErrComparable, left, right)
}

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
	case union:
		return evalUnion(e, curr, ctx)
	case intersect:
		return evalIntersect(e, curr, ctx)
	case except:
		return evalExcept(e, curr, ctx)
	case list:
		return evalList(e, curr, ctx)
	case root:
		return evalRoot(e, curr, ctx)
	case current:
		return evalCurrent(e, curr, ctx)
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
	case call:
		return evalCall(e, curr, ctx)
	default:
		return nil, ErrEval
	}
	return nil, nil
}

func evalList(expr list, curr Focus, ctx *Context) (Sequence, error) {
	var seq Sequence
	for i := range expr {
		res, err := eval(expr[i], curr, ctx)
		if err != nil {
			return nil, err
		}
		seq = append(seq, res...)
	}
	return seq, nil
}

func evalUnion(expr union, curr Focus, ctx *Context) (Sequence, error) {
	var seq Sequence
	for i := range expr {
		res, err := eval(expr[i], curr, ctx)
		if err != nil {
			return nil, err
		}
		seq = append(seq, res...)
	}
	return seq, nil
}

func evalExcept(expr except, curr Focus, ctx *Context) (Sequence, error) {
	var seq Sequence
	for i := range expr {
		res, err := eval(expr[i], curr, ctx)
		if err != nil {
			return nil, err
		}
		seq = append(seq, res...)
	}
	return seq, nil
}

func evalIntersect(expr intersect, curr Focus, ctx *Context) (Sequence, error) {
	var seq Sequence
	for i := range expr {
		res, err := eval(expr[i], curr, ctx)
		if err != nil {
			return nil, err
		}
		seq = append(seq, res...)
	}
	return seq, nil
}

func evalCall(expr call, curr Focus, ctx *Context) (Sequence, error) {
	return nil, nil
}

func evalPath(expr path, curr Focus, ctx *Context) (Sequence, error) {
	seq := Singleton(curr.Item)
	for i := range expr {
		var res Sequence
		for j := range seq {
			if seq[j].Atomic() {
				return nil, fmt.Errorf("%w not allow in path", ErrAtomic)
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

func evalCurrent(_ current, curr Focus, ctx *Context) (Sequence, error) {
	return Singleton(curr.Item), nil
}

func evalStep(expr step, curr Focus, ctx *Context) (Sequence, error) {
	switch expr.axis {
	default:
		return nil, fmt.Errorf("unsupported axis")
	case SelfAxis:
		if curr.Item.Atomic() {
			return nil, fmt.Errorf("%w not allowed with axis", ErrAtomic)
		}
		return eval(expr.Expr, curr.Sub(curr.Item, 1, 1), ctx)
	case AttributeAxis:
		node, err := currentNode(curr.Item)
		if err != nil {
			return nil, err
		}
		var (
			seq   Sequence
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
		ok, err := isMatch(sub, i+1)
		if err != nil {
			return nil, err
		}
		if ok {
			res = append(res, seq[i])
		}
	}
	return res, nil
}

func evalBinary(expr binary, curr Focus, ctx *Context) (Sequence, error) {
	switch expr.op {
	case Add, Sub, Mul, Div, Mod:
		return evalArithmetic(expr, curr, ctx)
	case Concat:
		return evalConcat(expr, curr, ctx)
	case Eq, Ne, Lt, Le, Gt, Ge:
		return evalComparison(expr, curr, ctx)
	case And:
		return evalAnd(expr, curr, ctx)
	case Or:
		return evalOr(expr, curr, ctx)
	default:
		return nil, fmt.Errorf("unsupported binary operator")
	}
}

func evalArithmetic(expr binary, curr Focus, ctx *Context) (Sequence, error) {
	var (
		value float64
		err   error
	)
	seq1, err := eval(expr.left, curr, ctx)
	if err != nil {
		return nil, err
	}
	seq2, err := eval(expr.right, curr, ctx)
	if err != nil {
		return nil, err
	}
	switch expr.op {
	case Add:
		value, err = doAdd(seq1, seq2)
	case Sub:
		value, err = doSub(seq1, seq2)
	case Mul:
		value, err = doMul(seq1, seq2)
	case Div:
		value, err = doDiv(seq1, seq2)
	case Mod:
		value, err = doMod(seq1, seq2)
	default:
		err = fmt.Errorf("unsupported arithmetic operator")
	}
	if err != nil {
		return nil, err
	}
	return Singleton(value), nil
}

func evalConcat(expr binary, curr Focus, ctx *Context) (Sequence, error) {
	seq1, err := eval(expr.left, curr, ctx)
	if err != nil {
		return nil, err
	}
	seq2, err := eval(expr.right, curr, ctx)
	if err != nil {
		return nil, err
	}

	if seq1.Len() > 1 || seq2.Len() > 1 {
		return nil, fmt.Errorf("sequences too long")
	}
	var left string
	if !seq1.Empty() {
		left, err = stringify(seq1[0].Value())
		if err != nil {
			return nil, err
		}
	}
	var right string
	if !seq2.Empty() {
		right, err = stringify(seq2[0].Value())
		if err != nil {
			return nil, err
		}
	}
	return Singleton(left + right), nil
}

func evalAnd(expr binary, curr Focus, ctx *Context) (Sequence, error) {
	left, err := eval(expr.left, curr, ctx)
	if err != nil {
		return nil, err
	}
	ok, err := effectiveBooleanValue(left)
	if err != nil || !ok {
		return Singleton(false), err
	}
	right, err := eval(expr.right, curr, ctx)
	if err != nil {
		return nil, err
	}
	ok, err = effectiveBooleanValue(right)
	if err != nil || !ok {
		return Singleton(false), err
	}
	return Singleton(true), nil
}

func evalOr(expr binary, curr Focus, ctx *Context) (Sequence, error) {
	left, err := eval(expr.left, curr, ctx)
	if err != nil {
		return nil, err
	}
	ok, err := effectiveBooleanValue(left)
	if err != nil || ok {
		return Singleton(true), err
	}
	right, err := eval(expr.right, curr, ctx)
	if err != nil {
		return nil, err
	}
	ok, err = effectiveBooleanValue(right)
	if err != nil || ok {
		return Singleton(true), err
	}
	return Singleton(false), nil
}

func evalComparison(expr binary, curr Focus, ctx *Context) (Sequence, error) {
	var (
		value any
		err   error
	)
	seq1, err := eval(expr.left, curr, ctx)
	if err != nil {
		return nil, err
	}
	seq2, err := eval(expr.right, curr, ctx)
	if err != nil {
		return nil, err
	}
	switch expr.op {
	case Eq:
		value, err = isEq(seq1, seq2)
	case Ne:
		value, err = isNe(seq1, seq2)
	case Lt:
		value, err = isLt(seq1, seq2)
	case Le:
		value, err = isLe(seq1, seq2)
	case Gt:
		value, err = isGt(seq1, seq2)
	case Ge:
		value, err = isGe(seq1, seq2)
	default:
		err = fmt.Errorf("unsupported comparison operator")
	}
	if err != nil {
		return nil, err
	}
	return Singleton(value), nil
}

func evalUnary(expr unary, curr Focus, ctx *Context) (Sequence, error) {
	seq, err := eval(expr.right, curr, ctx)
	if err != nil {
		return nil, err
	}
	switch expr.op {
	case Sub:
		if seq.Len() != 1 {
			return nil, fmt.Errorf("singleton sequence expected")
		}
		val, err := atomize(seq[0].Value())
		if err != nil {
			return nil, err
		}
		f, err := castToFloat(val)
		if err != nil {
			return nil, err
		}
		return Singleton(-f), nil
	default:
		return nil, fmt.Errorf("unsupported unary operator")
	}
	return nil, nil
}

func evalNumber(expr number, curr Focus, ctx *Context) (Sequence, error) {
	return Singleton(expr.value), nil
}

func evalLiteral(expr literal, curr Focus, ctx *Context) (Sequence, error) {
	return Singleton(expr.str), nil
}

func isEq(left, right Sequence) (bool, error) {
	return compare(left, right, func(left, right any) (bool, error) {
		return left == right, nil
	})
}

func isNe(left, right Sequence) (bool, error) {
	return compare(left, right, func(left, right any) (bool, error) {
		return left != right, nil
	})
}

func isGt(left, right Sequence) (bool, error) {
	return compare(left, right, func(left, right any) (bool, error) {
		switch e := left.(type) {
		case float64:
			r, ok := right.(float64)
			if !ok {
				return false, compareError(left, right)
			}
			return e > r, nil
		case string:
			r, ok := right.(string)
			if !ok {
				return false, compareError(left, right)
			}
			return e > r, nil
		default:
			return false, compareError(left, right)
		}
	})
}

func isGe(left, right Sequence) (bool, error) {
	return compare(left, right, func(left, right any) (bool, error) {
		switch e := left.(type) {
		case float64:
			r, ok := right.(float64)
			if !ok {
				return false, compareError(left, right)
			}
			return e >= r, nil
		case string:
			r, ok := right.(string)
			if !ok {
				return false, compareError(left, right)
			}
			return e >= r, nil
		default:
			return false, compareError(left, right)
		}
	})
}

func isLt(left, right Sequence) (bool, error) {
	return compare(left, right, func(left, right any) (bool, error) {
		switch e := left.(type) {
		case float64:
			r, ok := right.(float64)
			if !ok {
				return false, compareError(left, right)
			}
			return e < r, nil
		case string:
			r, ok := right.(string)
			if !ok {
				return false, compareError(left, right)
			}
			return e < r, nil
		default:
			return false, compareError(left, right)
		}
	})
}

func isLe(left, right Sequence) (bool, error) {
	return compare(left, right, func(left, right any) (bool, error) {
		switch e := left.(type) {
		case float64:
			r, ok := right.(float64)
			if !ok {
				return false, compareError(left, right)
			}
			return e <= r, nil
		case string:
			r, ok := right.(string)
			if !ok {
				return false, compareError(left, right)
			}
			return e <= r, nil
		default:
			return false, compareError(left, right)
		}
	})
}

func compare(left, right Sequence, cmp func(left, right any) (bool, error)) (bool, error) {
	if left.Empty() || right.Empty() {
		return false, nil
	}
	for i := range left {
		for j := range right {
			e, r, err := convertTo(left[i], right[j])
			if err != nil {
				return false, err
			}
			ok, err := cmp(e, r)
			if ok || err != nil {
				return ok, err
			}
		}
	}
	return false, nil
}

func doAdd(left, right Sequence) (float64, error) {
	return compute(left, right, func(left, right float64) (float64, error) {
		return left + right, nil
	})
}

func doSub(left, right Sequence) (float64, error) {
	return compute(left, right, func(left, right float64) (float64, error) {
		return left - right, nil
	})
}

func doMul(left, right Sequence) (float64, error) {
	return compute(left, right, func(left, right float64) (float64, error) {
		return left * right, nil
	})
}

func doDiv(left, right Sequence) (float64, error) {
	return compute(left, right, func(left, right float64) (float64, error) {
		if right == 0 {
			return 0, fmt.Errorf("division by zeri")
		}
		return left / right, nil
	})
}

func doMod(left, right Sequence) (float64, error) {
	return compute(left, right, func(left, right float64) (float64, error) {
		if right == 0 {
			return 0, fmt.Errorf("division by zeri")
		}
		return math.Mod(left, right), nil
	})
}

func compute(left, right Sequence, do func(left, right float64) (float64, error)) (float64, error) {
	if left.Empty() || right.Empty() {
		return 0, nil
	}
	if left.Len() != 1 || right.Len() != 1 {
		return 0, fmt.Errorf("singletons sequence expected on arithmetic")
	}
	x1, ok := left[0].Value().(float64)
	if !ok {
		return 0, fmt.Errorf("number expected on left side of arithmetic operation")
	}
	x2, ok := right[0].Value().(float64)
	if !ok {
		return 0, fmt.Errorf("number expected on right side of arithmetic operation")
	}
	return do(x1, x2)
}

func isMatch(seq Sequence, index int) (bool, error) {
	if seq.Singleton() && seq[0].Atomic() {
		n, ok := seq[0].Value().(float64)
		if ok {
			return int(n) == index, nil
		}
	}
	return effectiveBooleanValue(seq)
}

func effectiveBooleanValue(seq Sequence) (bool, error) {
	if seq.Empty() {
		return false, nil
	}

	if seq[0].Atomic() {
		if seq.Len() != 1 {
			return false, fmt.Errorf("ebv is undefined for a sequence of multiple atomic values")
		}

		switch x := seq[0].Value().(type) {
		case string:
			return x != "", nil
		case float64:
			return x != 0 && !math.IsNaN(x), nil
		case bool:
			return x, nil
		default:
			return false, fmt.Errorf("unsupported atomic type %T", x)
		}
	}

	for i := range seq {
		if seq[i].Atomic() {
			return false, fmt.Errorf("ebv is undefined for a mixed sequence")
		}
	}
	return true, nil
}

func currentNode(item Item) (xml.Node, error) {
	if item.Atomic() {
		return nil, fmt.Errorf("node item expected")
	}
	return item.Node(), nil
}

func traverseAttributes(node xml.Node) []*xml.Attribute {
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
	case xml.Attribute:
		return n.Name, true
	default:
		return xml.Name{}, false
	}
}
