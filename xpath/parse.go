package xpath

import (
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

func (p *Parser) Parse() error {
	return p.scan.Err()
}

func (p *Parser) done() bool {
	return p.is(Eof)
}

func (p *Parser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}
