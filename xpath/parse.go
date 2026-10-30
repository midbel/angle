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

	infix   map[rune]func(Expr) (Expr, error)
	postfix map[rune]func(Expr) (Expr, error)
	prefix  map[rune]func() (Expr, error)
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
		Slash:    p.parseSlash,
		Deep:     p.parseDeep,
		Attr:     p.parseAttribute,
		Name:     p.parseName,
		Mul:      p.parseName,
		Current:  p.parseCurrent,
		Parent:   p.parseParent,
	}
	p.postfix = map[rune]func(Expr) (Expr, error){
		BegGrp:  p.parseCall,
		BegPred: p.parsePredicate,
		Slash:   p.parseStep,
		Deep:    p.parseStep,
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
		Comma:  p.parseList,
		Pipe:   p.parseUnion,
	}
	p.next()
	p.next()
	return p
}

func (p *Parser) Parse() (Expr, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if !p.done() {
		return nil, fmt.Errorf("eof expected")
	}
	return expr, p.scan.Err()
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.startExpr(powLowest)
}

func (p *Parser) startExpr(pow int) (Expr, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported prefix operator")
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for {
		fn, ok := p.postfix[p.curr.Type]
		if !ok {
			break
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	for !p.done() && pow < p.powerOf() {
		fn, ok := p.infix[p.curr.Type]
		if !ok {
			return nil, fmt.Errorf("unsupported prefix operator")
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parseCall(left Expr) (Expr, error) {
	n, ok := left.(name)
	if !ok {
		return nil, fmt.Errorf("name expected for function")
	}
	p.next()
	args, err := p.parseArgs()
	if err != nil {
		return nil, err
	}
	if !p.is(EndGrp) {
		return nil, p.unexpectedToken()
	}
	expr := call{
		Name: n,
		Args: args,
	}
	return expr, nil
}

func (p *Parser) parseGroup() (Expr, error) {
	p.next()
	expr, err := p.startExpr(powGroup)
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

func (p *Parser) parsePredicate(left Expr) (Expr, error) {
	return nil, nil
}

func (p *Parser) parseSlash() (Expr, error) {
	return nil, nil
}

func (p *Parser) parseDeep() (Expr, error) {
	return nil, nil
}

func (p *Parser) parseStep(left Expr) (Expr, error) {
	return nil, nil
}

func (p *Parser) parseList(left Expr) (Expr, error) {
	var expr []Expr
	for p.is(Comma) {
		p.next()
		e, err := p.startExpr(powUnion)
		if err != nil {
			return nil, err
		}
		expr = append(expr, e)
	}
	return list(expr), nil
}

func (p *Parser) parseUnion(left Expr) (Expr, error) {
	var expr []Expr
	for p.is(Pipe) {
		p.next()
		e, err := p.startExpr(powUnion)
		if err != nil {
			return nil, err
		}
		expr = append(expr, e)
	}
	return union(expr), nil
}

func (p *Parser) parseName() (Expr, error) {
	n := name{
		Name: p.currentLiteral(),
	}
	p.next()
	if p.is(Namespace) {
		n.Prefix = n.Name
		p.next()
		if !p.is(Mul) && !p.is(Name) {
			return nil, p.unexpectedToken()
		}
		n.Name = p.currentLiteral()
		p.next()
	}
	return n, nil
}

func (p *Parser) parseAttribute() (Expr, error) {
	return nil, nil
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

func (p *Parser) parseArgs() ([]Expr, error) {
	var list []Expr
	for !p.done() && !p.is(EndGrp) {
		a, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		list = append(list, a)
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
	return list, nil
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
