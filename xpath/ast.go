package xpath

import (
	"fmt"
	"strings"
)

type Expr interface {
	fmt.Stringer
	expr()
}

type Path []Expr

func (p Path) String() string {
	var str strings.Builder
	str.WriteString("path(")
	for i := range p {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(p[i].String())
	}
	str.WriteString(")")
	return str.String()
}

func (Path) expr() {}

type Union []Expr

func (u Union) String() string {
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

func (Union) expr() {}

type List []Expr

func (i List) String() string {
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

func (List) expr() {}

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

type Step struct {
	axis Axis
	Expr
}

func (s Step) String() string {
	return fmt.Sprintf("step(axis: %s, expr: %s)", s.axis, s.Expr)
}

func (Step) expr() {}

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

type nameTest struct {
	Prefix string
	Name   string
}

func (n nameTest) String() string {
	return fmt.Sprintf("nameTest(prefix: %s, name: %s)", n.Prefix, n.Name)
}

func (nameTest) expr() {}

type predicate struct {
	Expr
	Check Expr
}

func (p predicate) String() string {
	return fmt.Sprintf("predicate(expr: %s, check: %s", p.Expr, p.Check)
}

func (p predicate) expr() {}

func axisStep(axis Axis, expr Expr) Expr {
	return Step{
		axis: axis,
		Expr: expr,
	}
}

func descendantOrSelfStep() Expr {
	return Step{
		axis: DescendantOrSelfAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}

func parentStep() Expr {
	return Step{
		axis: SelfAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}

func currentStep() Expr {
	return Step{
		axis: ParentAxis,
		Expr: typeTest{
			Kind: NodeTypeNode,
		},
	}
}
