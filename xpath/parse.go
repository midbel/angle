package xpath

import (
	"fmt"
	"io"
	"strconv"
)

type Parser struct {
	scan *Scanner
	curr Token
	peek Token

	infix  map[rune]func(Expr) (Expr, error)
	prefix map[rune]func() (Expr, error)
}

func NewParser(r io.Reader) *Parser {
	p := &Parser{
		scan: Scan(r),
	}
	p.prefix = map[rune]func() (Expr, error){
		Sub:      p.parseReverse,
		BegGrp:   p.parseGroup,
		Literal:  p.parseLiteral,
		Number:   p.parseNumber,
		Variable: p.parseVariable,
	}
	p.infix = map[rune]func(Expr) (Expr, error){
		Eq:     p.parseBinary,
		Ne:     p.parseBinary,
		Lt:     p.parseBinary,
		Le:     p.parseBinary,
		Gt:     p.parseBinary,
		Ge:     p.parseBinary,
		Add:    p.parseBinary,
		Sub:    p.parseBinary,
		Mul:    p.parseBinary,
		Div:    p.parseBinary,
		Mod:    p.parseBinary,
		Concat: p.parseBinary,
		And:    p.parseBinary,
		Or:     p.parseBinary,
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

func (p *Parser) parseExpr() (Expr, error) {
	return p.startExpr(powLowest)
}

func (p *Parser) startExpr(pow int) (Expr, error) {
	left, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}
	for !p.done() && !p.endOfExpr() && pow < p.powerOf() {
		left, err = p.parseInfix(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parsePrefix() (Expr, error) {
	fn, ok := p.prefix[p.curr.Type]
	if ok {
		return fn()
	}
	return p.parsePath()
}

func (p *Parser) parseInfix(left Expr) (Expr, error) {
	fn, ok := p.infix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported infix operator")
	}
	return fn(left)
}

func (p *Parser) parseGroup() (Expr, error) {
	p.next()
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if !p.is(EndGrp) {
		return nil, p.unexpectedToken()
	}
	p.next()
	return expr, nil
}

func (p *Parser) parseVariable() (Expr, error) {
	v := variable{
		ident: p.currentLiteral(),
	}
	p.next()
	return v, nil
}

func (p *Parser) parseNumber() (Expr, error) {
	v, err := strconv.ParseFloat(p.currentLiteral(), 64)
	if err != nil {
		return nil, err
	}
	n := number{
		value: v,
	}
	p.next()
	return n, nil
}

func (p *Parser) parseLiteral() (Expr, error) {
	i := literal{
		str: p.currentLiteral(),
	}
	p.next()
	return i, nil
}

func (p *Parser) parseReverse() (Expr, error) {
	u := unary{
		op: p.curr.Type,
	}
	p.next()
	expr, err := p.startExpr(powPrefix)
	if err != nil {
		return nil, err
	}
	u.right = expr
	return u, nil
}

func (p *Parser) parseBinary(left Expr) (Expr, error) {
	b := binary{
		left: left,
		op:   p.curr.Type,
	}
	pow := p.powerOf()
	p.next()
	right, err := p.startExpr(pow)
	if err != nil {
		return nil, err
	}
	b.right = right
	return b, nil
}

func (p *Parser) parseList() (Expr, error) {
	var list []Expr
	for !p.endOfExpr() {
		expr, err := p.parseUnion()
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
		case p.is(EndGrp):
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

func (p *Parser) parseUnion() (Expr, error) {
	var list []Expr
	for !p.done() {
		expr, err := p.parsePath()
		if err != nil {
			return nil, err
		}
		list = append(list, expr)
		if !p.is(Pipe) {
			break
		}
		p.next()
	}
	if len(list) == 1 {
		return list[0], nil
	}
	return Union(list), nil
}

func (p *Parser) parseNestedPath() (Expr, error) {
	p.next()
	expr, err := p.parseList()
	if err != nil {
		return nil, err
	}
	if !p.is(EndGrp) {
		return nil, p.unexpectedToken()
	}
	p.next()
	return expr, nil
}

func (p *Parser) parsePath() (Expr, error) {
	var (
		list  []Expr
		start Expr
		err   error
	)
	switch {
	case p.is(BegGrp):
		start, err = p.parseNestedPath()
	case p.is(Current):
		start, err = p.parseCurrent()
	case p.is(Parent):
		start, err = p.parseParent()
	case p.is(Name) || p.is(Mul):
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
	} else if p.is(BegGrp) {
		expr, err := p.parseNestedPath()
		if err != nil {
			return nil, err
		}
		return append(list, expr), nil
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
	case p.is(Name) || p.is(Mul):
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
		return p.parseParens(nt.Name)
	}
	if p.is(Namespace) {
		nt.Prefix = nt.Name
		p.next()
		if !p.is(Mul) && !p.is(Name) {
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

func (p *Parser) parseParens(name string) (Expr, error) {
	p.next()
	var (
		kind NodeType
		expr Expr
		err  error
	)
	switch name {
	case "node":
		kind = NodeTypeNode
	case "text":
		kind = NodeTypeText
	case "comment":
		kind = NodeTypeComment
	case "processing-instruction":
		kind = NodeTypePI
	default:
	}
	if kind == 0 {
		expr, err = p.parseCall(name)
		if err != nil {
			return nil, err
		}
	} else {
		expr = typeTest{
			Kind: kind,
		}
	}
	if !p.is(EndGrp) {
		return nil, p.unexpectedToken()
	}
	p.next()
	return expr, nil
}

func (p *Parser) parseCall(name string) (Expr, error) {
	expr := call{
		Name: name,
	}
	for !p.done() && !p.is(EndGrp) {
		a, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		expr.Args = append(expr.Args, a)
		switch {
		case p.is(EndGrp):
		case p.is(Comma):
			p.next()
			if p.is(EndGrp) {
				return nil, fmt.Errorf("invalid syntax, expected expression after comma")
			}
		default:
			return nil, p.unexpectedToken()
		}
	}
	return expr, nil
}

func (p *Parser) parseCurrent() (Expr, error) {
	p.next()
	return currentStep(), nil
}

func (p *Parser) parseParent() (Expr, error) {
	p.next()
	return parentStep(), nil
}

func (p *Parser) done() bool {
	return p.is(Eof)
}

func (p *Parser) currentLiteral() string {
	if p.is(Mul) {
		return "*"
	}
	return p.curr.Literal
}

func (p *Parser) endOfExpr() bool {
	return p.done() || p.is(Comma) || p.is(Pipe) || p.is(EndGrp)
}

func (p *Parser) peekEndOfExpr() bool {
	return p.peekIs(Eof) || p.peekIs(Comma) || p.peekIs(Pipe) || p.peekIs(EndGrp)
}

func (p *Parser) peekIs(kind rune) bool {
	return p.peek.Type == kind
}

func (p *Parser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *Parser) powerOf() int {
	pow, ok := bindings[p.curr.Type]
	if !ok {
		pow = powLowest
	}
	return pow
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *Parser) unexpectedToken() error {
	return fmt.Errorf("%s: unexpected token: %s", p.curr.Position, p.curr)
}
