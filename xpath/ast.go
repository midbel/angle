package xpath

import (
	"fmt"
	"strings"
)

type Expr interface {
	fmt.Stringer
	expr()
}

type path []Expr

func createPath(ps ...Expr) Expr {
	return path(ps)
}

func (ps path) String() string {
	var str strings.Builder
	str.WriteString("path(")
	for i := range ps {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(ps[i].String())
	}
	str.WriteString(")")
	return str.String()
}

func (path) expr() {}

type union struct {
	Left  Expr
	Right Expr
}

func createUnion(left, right Expr) Expr {
	return union{
		Left:  left,
		Right: right,
	}
}

func (us union) String() string {
	var str strings.Builder
	str.WriteString("union(")
	str.WriteString(us.Left.String())
	str.WriteString(", ")
	str.WriteString(us.Right.String())
	str.WriteString(")")
	return str.String()
}

func (union) expr() {}

type intersect struct {
	Left  Expr
	Right Expr
}

func createIntersect(left, right Expr) Expr {
	return intersect{
		Left:  left,
		Right: right,
	}
}

func (is intersect) String() string {
	var str strings.Builder
	str.WriteString("intersect(")
	str.WriteString(is.Left.String())
	str.WriteString(", ")
	str.WriteString(is.Right.String())
	str.WriteString(")")
	return str.String()
}

func (intersect) expr() {}

type except struct {
	Left  Expr
	Right Expr
}

func createExcept(left, right Expr) Expr {
	return except{
		Left:  left,
		Right: right,
	}
}

func (es except) String() string {
	var str strings.Builder
	str.WriteString("except(")
	str.WriteString(es.Left.String())
	str.WriteString(", ")
	str.WriteString(es.Right.String())
	str.WriteString(")")
	return str.String()
}

func (except) expr() {}

type list []Expr

func createList(is ...Expr) Expr {
	return list(is)
}

func (is list) String() string {
	var str strings.Builder
	str.WriteString("list(")
	for j := range is {
		if j > 0 {
			str.WriteString(", ")
		}
		str.WriteString(is[j].String())
	}
	str.WriteString(")")
	return str.String()
}

func (list) expr() {}

type root struct{}

func createRoot() Expr {
	return root{}
}

func (root) String() string {
	return "root"
}

func (root) expr() {}

type current struct{}

func (current) String() string {
	return "current"
}

func (current) expr() {}

type Axis uint8

const (
	ChildAxis Axis = 1 + iota
	SelfAxis
	DescendantAxis
	DescendantOrSelfAxis
	ParentAxis
	AncestorAxis
	AncestorOrSelfAxis
	FollowingAxis
	FollowingSiblingAxis
	PrecedingAxis
	PrecedingSiblingAxis
	AttributeAxis
)

func (a Axis) String() string {
	switch a {
	default:
		return "unknown-axis"
	case ChildAxis:
		return "child-axis"
	case DescendantAxis:
		return "descendant-axis"
	case ParentAxis:
		return "parent-axis"
	case AncestorAxis:
		return "ancestor-axis"
	case FollowingSiblingAxis:
		return "following-sibling-axis"
	case PrecedingSiblingAxis:
		return "preceding-sibling-axis"
	case FollowingAxis:
		return "following-axis"
	case PrecedingAxis:
		return "preceding-axis"
	case AttributeAxis:
		return "attribute-axis"
	case SelfAxis:
		return "self-axis"
	case DescendantOrSelfAxis:
		return "descendant-or-self-axis"
	case AncestorOrSelfAxis:
		return "ancestor-or-self-axis"
	}
}

type step struct {
	axis Axis
	Expr
}

func (s step) String() string {
	return fmt.Sprintf("step(axis: %s, expr: %s)", s.axis, s.Expr)
}

func (step) expr() {}

type NodeType uint8

const (
	NodeTypeNode NodeType = 1 + iota
	NodeTypeElem
	NodeTypeAttr
	NodeTypeText
	NodeTypeComment
	NodeTypePI
)

func (n NodeType) String() string {
	switch n {
	default:
		return "unknown-type"
	case NodeTypeNode:
		return "node"
	case NodeTypeElem:
		return "element-node"
	case NodeTypeAttr:
		return "attribute-node"
	case NodeTypeText:
		return "text-node"
	case NodeTypeComment:
		return "comment-node"
	case NodeTypePI:
		return "pi-node"
	}
}

type typeTest struct {
	Kind NodeType
}

func (t typeTest) String() string {
	return fmt.Sprintf("type(%d)", t.Kind)
}

func (typeTest) expr() {}

type nameTest struct {
	Prefix string
	Name   string
}

func createName(prefix, name string) Expr {
	return nameTest{
		Prefix: prefix,
		Name:   name,
	}
}

func (n nameTest) String() string {
	return fmt.Sprintf("name (prefix: %s, name: %s)", n.Prefix, n.Name)
}

func (nameTest) expr() {}

type predicate struct {
	Expr
	Check Expr
}

func createPredicate(expr, check Expr) Expr {
	return predicate{
		Expr:  expr,
		Check: check,
	}
}

func (p predicate) String() string {
	return fmt.Sprintf("predicate(expr: %s, check: %s)", p.Expr, p.Check)
}

func (p predicate) expr() {}

type call struct {
	Name nameTest
	Args []Expr
}

func (c call) String() string {
	var str strings.Builder
	str.WriteString("call(name: ")
	str.WriteString(c.Name.String())
	str.WriteString(", args: ")
	for i := range c.Args {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(c.Args[i].String())
	}
	str.WriteString(")")
	return str.String()
}

func (call) expr() {}

type literal struct {
	str string
}

func createLiteral(str string) Expr {
	return literal{
		str: str,
	}
}

func (i literal) String() string {
	return fmt.Sprintf("literal(%s)", i.str)
}

func (literal) expr() {}

type variable struct {
	ident string
}

func createVariable(ident string) Expr {
	return variable{
		ident: ident,
	}
}

func (v variable) String() string {
	return fmt.Sprintf("variable(%s)", v.ident)
}

func (variable) expr() {}

type number struct {
	value float64
}

func createNumber(value float64) Expr {
	return number{
		value: value,
	}
}

func (n number) String() string {
	return fmt.Sprintf("number(%f)", n.value)
}

func (number) expr() {}

type binary struct {
	left  Expr
	right Expr
	op    rune
}

func createBinary(left, right Expr, op rune) Expr {
	return binary{
		left:  left,
		right: right,
		op:    op,
	}
}

func (b binary) String() string {
	return fmt.Sprintf("binary(left: %s, right: %s, op: %c)", b.left, b.right, b.op)
}

func (binary) expr() {}

type unary struct {
	right Expr
	op    rune
}

func createUnary(expr Expr, op rune) Expr {
	return unary{
		right: expr,
		op:    op,
	}
}

func (u unary) String() string {
	return fmt.Sprintf("unary(expr: %s, op: %c)", u.right, u.op)
}

func (unary) expr() {}

type assign struct {
	Ident string
	Expr  Expr
}

func createAssignment(ident string, expr Expr) assign {
	return assign{
		Ident: ident,
		Expr: expr,
	}
}

type some struct{}

func (s some) String() string {
	return "some"
}

func (some) expr() {}

type every struct{}

func (e every) String() string {
	return "every"
}

func (every) expr() {}

type let struct {
	Values []assign
	Body   Expr
}

func (e let) String() string {
	var str strings.Builder
	str.WriteString("let(assignment: ")
	for i := range e.Values {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(e.Values[i].Ident)
		str.WriteString("=")
		str.WriteString(e.Values[i].Expr.String())
	}
	str.WriteString(", body: ")
	str.WriteString(e.Body.String())
	str.WriteString(")")
	return str.String()
}

func (let) expr() {}

type conditional struct {
	Test Expr
	Csq  Expr
	Alt  Expr
}

func (c conditional) String() string {
	var str strings.Builder
	str.WriteString("if(test:")
	str.WriteString(c.Test.String())
	str.WriteString(", csq: ")
	str.WriteString(c.Csq.String())
	str.WriteString(", alt: ")
	str.WriteString(c.Alt.String())
	str.WriteString(")")
	return str.String()
}

func (conditional) expr() {}

type loop struct {
	Values []assign
	Body   Expr
}

func (o loop) String() string {
	var str strings.Builder
	str.WriteString("for(clause: ")
	for i := range o.Values {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(o.Values[i].Ident)
		str.WriteString("=")
		str.WriteString(o.Values[i].Expr.String())
	}
	str.WriteString(", body: ")
	str.WriteString(o.Body.String())
	str.WriteString(")")
	return str.String()
}

func (loop) expr() {}

func axisStep(axis Axis, expr Expr) Expr {
	return step{
		axis: axis,
		Expr: expr,
	}
}

func attributeStep(expr Expr) Expr {
	return step{
		axis: AttributeAxis,
		Expr: expr,
	}
}

func descendantOrSelfStep() Expr {
	return step{
		axis: DescendantOrSelfAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}

func parentStep() Expr {
	return step{
		axis: ParentAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}

func currentStep() Expr {
	return step{
		axis: SelfAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}
