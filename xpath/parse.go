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
		Slash:    p.parseRoot,
		Deep:     p.parseRoot,
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
	fn, err := p.getPrefixFn()
	if err != nil {
		return nil, err
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for {
		fn, err := p.getPostfixFn()
		if err != nil {
			break
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	for !p.done() && pow < p.powerOf() {
		fn, err := p.getInfixFn()
		if err != nil {
			return nil, err
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) getPrefixFn() (func() (Expr, error), error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported prefix operator")
	}
	return fn, nil
}

func (p *Parser) getPostfixFn() (func(Expr) (Expr, error), error) {
	fn, ok := p.postfix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported prefix operator")
	}
	return fn, nil
}

func (p *Parser) getInfixFn() (func(Expr) (Expr, error), error) {
	if p.is(Name) {
		var kind rune
		switch p.currentLiteral() {
		case kwMod:
			kind = Mod
		case kwDiv:
			kind = Div
		case kwAnd:
			kind = And
		case kwOr:
			kind = Or
		default:
			return nil, fmt.Errorf("unsupported infix operator")
		}
		p.curr.Type = kind
		return p.parseBinary, nil
	}
	fn, ok := p.infix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported infix operator")
	}
	return fn, nil
}

func (p *Parser) parseNodeTest() (Expr, bool, error) {
	if !p.peekIs(BegGrp) {
		return nil, false, nil
	}
	var kind NodeType
	switch p.currentLiteral() {
	case "node":
		kind = NodeTypeNode
	case "element":
		kind = NodeTypeElem
	case "attribute":
		kind = NodeTypeAttr
	case "text":
		kind = NodeTypeText
	case "comment":
		kind = NodeTypeComment
	case "processing-instruction":
		kind = NodeTypePI
	default:
		return nil, false, nil
	}
	p.next()
	p.next()
	if !p.is(EndGrp) {
		return nil, true, fmt.Errorf("kind: %s", p.unexpectedToken())
	}
	p.next()
	nt := typeTest{
		Kind: kind,
	}
	return nt, true, nil
}

func (p *Parser) parseCall(left Expr) (Expr, error) {
	n, ok := left.(nameTest)
	if !ok {
		return nil, fmt.Errorf("name expected for function")
	}
	p.next()
	args, err := p.parseArgs()
	if err != nil {
		return nil, err
	}
	if !p.is(EndGrp) {
		return nil, fmt.Errorf("call: %s", p.unexpectedToken())
	}
	p.next()
	expr := call{
		Name: n,
		Args: args,
	}
	return expr, nil
}

func (p *Parser) parseGroup() (Expr, error) {
	p.next()
	expr, err := p.startExpr(powLowest)
	if err != nil {
		return nil, err
	}
	if !p.is(EndGrp) {
		return nil, fmt.Errorf("group: %s", p.unexpectedToken())
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
	p.next()
	expr, err := p.startExpr(powLowest)
	if err != nil {
		return nil, err
	}
	if !p.is(EndPred) {
		return nil, p.unexpectedToken()
	}
	p.next()
	pt := predicate{
		Expr:  left,
		Check: expr,
	}
	return pt, nil
}

func (p *Parser) parseRoot() (Expr, error) {
	var r root
	return p.parseStep(r)
}

func (p *Parser) parseStep(left Expr) (Expr, error) {
	var expr []Expr
	expr = append(expr, left)
	if p.is(Deep) {
		expr = append(expr, descendantOrSelfStep())
	}
	p.next()
	next, err := p.startExpr(powPath)
	if err != nil {
		return nil, err
	}
	if ps, ok := next.(path); ok {
		expr = append(expr, ps...)
	} else {
		expr = append(expr, next)
	}
	return path(expr), nil
}

func (p *Parser) parseList(left Expr) (Expr, error) {
	var expr []Expr
	expr = append(expr, left)
	fmt.Println("startList", p.curr, p.peek)
	for p.is(Comma) {
		p.next()
		fmt.Println("parseList", p.curr, p.peek)
		e, err := p.startExpr(powLowest)
		if err != nil {
			return nil, err
		}
		expr = append(expr, e)
	}
	return list(expr), nil
}

func (p *Parser) parseUnion(left Expr) (Expr, error) {
	var expr []Expr
	expr = append(expr, left)
	for p.is(Pipe) {
		p.next()
		e, err := p.startExpr(powLowest)
		if err != nil {
			return nil, err
		}
		expr = append(expr, e)
	}
	return union(expr), nil
}

func (p *Parser) parseName() (Expr, error) {
	var (
		st  step
		nt  nameTest
		err error
	)
	st.axis, err = p.parseAxis()
	if err != nil {
		return nil, err
	}
	if p.peekIs(Namespace) {
		if !p.is(Mul) && !p.is(Name) {
			return nil, p.unexpectedToken()
		}
		nt.Prefix = p.currentLiteral()
		p.next()
		p.next()
	}
	if !p.is(Mul) && !p.is(Name) {
		return nil, fmt.Errorf("name: %s", p.unexpectedToken())
	}
	if nt.Prefix == "" {
		if expr, ok, err := p.parseNodeTest(); ok {
			if err != nil {
				return nil, err
			}
			if st.axis == 0 {
				st.axis = ChildAxis
			}
			st.Expr = expr
			return st, nil
		}
	}
	nt.Name = p.currentLiteral()
	p.next()
	if p.is(BegGrp) {
		if st.axis != 0 {
			return nil, fmt.Errorf("function call can not used an axis")
		}
		return p.parseCall(nt)
	}
	if st.axis == 0 {
		st.axis = ChildAxis
	}
	st.Expr = nt
	return st, nil
}

func (p *Parser) parseAttribute() (Expr, error) {
	p.next()
	var nt nameTest
	if p.peekIs(Namespace) {
		if !p.is(Mul) && !p.is(Name) {
			return nil, fmt.Errorf("attribute: %s", p.unexpectedToken())
		}
		nt.Prefix = p.currentLiteral()
		p.next()
		p.next()
	}
	if !p.is(Mul) && !p.is(Name) {
		return nil, fmt.Errorf("attribute: %s", p.unexpectedToken())
	}
	nt.Name = p.currentLiteral()
	p.next()
	return attributeStep(nt), nil
}

func (p *Parser) parseAxis() (Axis, error) {
	if !p.peekIs(AxisSep) {
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
			return nil, fmt.Errorf("arg: %s", p.unexpectedToken())
		}
	}
	return list, nil
}

func (p *Parser) parseCurrent() (Expr, error) {
	p.next()
	return current{}, nil
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
	kind := p.curr.Type
	if p.is(Name) {
		switch p.currentLiteral() {
		case kwMod:
			kind = Mod
		case kwDiv:
			kind = Div
		case kwAnd:
			kind = And
		case kwOr:
			kind = Or
		default:
		}
	}
	pow, ok := bindings[kind]
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
