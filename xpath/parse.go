package xpath

import (
	"fmt"
	"io"
)

type Parser struct {
	scan *Scanner
	curr Token
	peek Token
}

func NewParser(r io.Reader) *Parser {
	p := &Parser{
		scan: Scan(r),
	}
	p.next()
	p.next()
	return p
}

func (p *Parser) Parse() (Expr, error) {
	path, err := p.parseList()
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, fmt.Errorf("eof expected")
	}
	return path, p.scan.Err()
}

func (p *Parser) parseList() (Expr, error) {
	var list []Expr
	for !p.done() {
		expr, err := p.parsePath()
		if err != nil {
			return nil, err
		}
		switch {
		case p.done():
		case p.is(Comma):
			p.next()
			if p.done() {
				return nil, fmt.Errorf("unexpected end of expression after comma")
			}
		default:
			return nil, p.unexpectedToken()
		}
		list = append(list, expr)
	}
	if len(list) == 1 {
		return list[0], nil
	}
	return List(list), nil
}

func (p *Parser) parsePath() (Expr, error) {
	var (
		list  []Expr
		start Expr
		err   error
	)
	switch {
	case p.is(Current):
		start, err = p.parseCurrent()
	case p.is(Parent):
		start, err = p.parseParent()
	case p.is(Name) || p.is(Wildcard):
		start, err = p.parseName()
	case p.is(Slash):
		if p.peekEndOfExpr() {
			p.next()
			return Root{}, nil
		}
	case p.is(Deep):
		if p.peekEndOfExpr() {
			start = descendantOrSelfStep()
			p.next()
			return start, nil
		}
	default:
		return nil, p.unexpectedToken()
	}
	if err != nil {
		return nil, err
	}
	if start != nil {
		list = append(list, start)
	}
	for !p.endOfExpr() {
		steps, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		list = append(list, steps...)
	}
	return Path(list), nil
}

func (p *Parser) parseStep() ([]Expr, error) {
	var list []Expr
	if !p.is(Slash) && !p.is(Deep) {
		return nil, p.unexpectedToken()
	}
	if p.is(Deep) {
		step := descendantOrSelfStep()
		list = append(list, step)
	}
	p.next()
	if p.done() {
		return nil, fmt.Errorf("invalid syntax, expected expression after '/' or '//'")
	}
	if p.is(Current) {
		step, err := p.parseCurrent()
		if err != nil {
			return nil, err
		}
		return append(list, step), nil
	} else if p.is(Parent) {
		step, err := p.parseParent()
		if err != nil {
			return nil, err
		}
		return append(list, step), nil
	}
	var (
		axis Axis
		step Expr
		err  error
	)
	axis, err = p.parseAxis()
	if err != nil {
		return nil, err
	}
	switch {
	case p.is(Name) || p.is(Wildcard):
		if axis == 0 {
			axis = ChildAxis
		}
		expr, err := p.parseName()
		if err != nil {
			return nil, err
		}
		step = axisStep(axis, expr)
	case p.is(Attr):
		if axis != 0 {
			return nil, fmt.Errorf("attribute: invalid syntax")
		}
		p.next()
		expr, err := p.parseName()
		if err != nil {
			return nil, err
		}
		step = axisStep(AttributeAxis, expr)
	default:
		return nil, p.unexpectedToken()
	}
	return append(list, step), err
}

func (p *Parser) parseName() (Expr, error) {
	nt := nameTest{
		Name: p.currentLiteral(),
	}
	p.next()
	if p.is(BegGrp) {
		return p.parseType(nt.Name)
	}
	if p.is(Namespace) {
		nt.Prefix = nt.Name
		p.next()
		if !p.is(Wildcard) && !p.is(Name) {
			return nil, p.unexpectedToken()
		}
		nt.Name = p.currentLiteral()
		p.next()
	}
	return nt, nil
}

func (p *Parser) parseAxis() (Axis, error) {
	if p.peek.Type != AxisSep {
		return 0, nil
	}
	if !p.is(Name) {
		return 0, p.unexpectedToken()
	}
	var axis Axis
	switch p.currentLiteral() {
	case "child":
		axis = ChildAxis
	case "descendant":
		axis = DescendantAxis
	case "ancestor":
		axis = AncestorAxis
	case "following-sibling":
		axis = FollowingSiblingAxis
	case "preceding-sibling":
		axis = PrecedingSiblingAxis
	case "following":
		axis = FollowingAxis
	case "preceding":
		axis = PrecedingAxis
	case "attribute":
		axis = AttributeAxis
	case "self":
		axis = SelfAxis
	case "descendant-or-self":
		axis = DescendantOrSelfAxis
	case "ancestor-or-self":
		axis = AncestorOrSelfAxis
	default:
		return 0, fmt.Errorf("invalid axis: %s", p.currentLiteral())
	}
	p.next()
	p.next()
	return axis, nil
}

func (p *Parser) parseType(name string) (Expr, error) {
	p.next()
	var tt typeTest
	switch name {
	case "node":
		tt.Kind = NodeTypeNode
	case "text":
		tt.Kind = NodeTypeText
	case "comment":
		tt.Kind = NodeTypeComment
	case "processing-instruction":
		tt.Kind = NodeTypePI
	default:
		return nil, fmt.Errorf("invalid type: %s", name)
	}
	if !p.is(EndGrp) {
		return nil, p.unexpectedToken()
	}
	p.next()
	return tt, nil
}

func (p *Parser) parseCurrent() (Expr, error) {
	p.next()
	return parentStep(), nil
}

func (p *Parser) parseParent() (Expr, error) {
	p.next()
	return currentStep(), nil
}

func (p *Parser) done() bool {
	return p.is(Eof)
}

func (p *Parser) currentLiteral() string {
	if p.is(Wildcard) {
		return "*"
	}
	return p.curr.Literal
}

func (p *Parser) endOfExpr() bool {
	return p.done() || p.is(Comma) || p.is(Pipe)
}

func (p *Parser) peekEndOfExpr() bool {
	return p.peekIs(Eof) || p.peekIs(Comma) || p.peekIs(Pipe)
}

func (p *Parser) peekIs(kind rune) bool {
	return p.peek.Type == kind
}

func (p *Parser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *Parser) unexpectedToken() error {
	return fmt.Errorf("unexpected token: %s", p.curr)
}
