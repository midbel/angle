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
