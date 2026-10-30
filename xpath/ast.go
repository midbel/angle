package xpath

import (
	"fmt"
	"strings"
)

type Expr interface {
	fmt.Stringer
	expr()
}

type union []Expr

func (u union) String() string {
	var str strings.Builder
	str.WriteString("union(")
	for i := range u {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(u[i].String())
	}
	str.WriteString(")")
	return str.String()
}

func (union) expr() {}

type list []Expr

func (i list) String() string {
	var str strings.Builder
	str.WriteString("list(")
	for j := range i {
		if j > 0 {
			str.WriteString(", ")
		}
		str.WriteString(i[j].String())
	}
	str.WriteString(")")
	return str.String()
}

func (list) expr() {}

type Root struct{}

func (Root) String() string {
	return "root"
}

func (Root) expr() {}

type Axis uint8

const (
	ChildAxis Axis = 1 + iota
	DescendantAxis
	ParentAxis
	AncestorAxis
	FollowingSiblingAxis
	PrecedingSiblingAxis
	FollowingAxis
	PrecedingAxis
	AttributeAxis
	SelfAxis
	DescendantOrSelfAxis
	AncestorOrSelfAxis
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

type name struct {
	Prefix string
	Name   string
}

func (n name) String() string {
	return fmt.Sprintf("nameTest(prefix: %s, name: %s)", n.Prefix, n.Name)
}

func (name) expr() {}

type predicate struct {
	Expr
	Check Expr
}

func (p predicate) String() string {
	return fmt.Sprintf("predicate(expr: %s, check: %s)", p.Expr, p.Check)
}

func (p predicate) expr() {}

type call struct {
	Name name
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

func (i literal) String() string {
	return fmt.Sprintf("literal(%s)", i.str)
}

func (literal) expr() {}

type variable struct {
	ident string
}

func (v variable) String() string {
	return fmt.Sprintf("variable(%s)", v.ident)
}

func (variable) expr() {}

type number struct {
	value float64
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

func (b binary) String() string {
	return fmt.Sprintf("binary(left: %s, right: %s, op: %c)", b.left, b.right, b.op)
}

func (binary) expr() {}

type unary struct {
	right Expr
	op    rune
}

func (u unary) String() string {
	return fmt.Sprintf("unary(expr: %s, op: %c)", u.right, u.op)
}

func (unary) expr() {}

func axisStep(axis Axis, expr Expr) Expr {
	return step{
		axis: axis,
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
		axis: SelfAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}

func currentStep() Expr {
	return step{
		axis: ParentAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}
