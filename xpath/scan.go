package xpath

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Position struct {
	Line   int
	Column int
}

type Token struct {
	Literal string
	Type    rune
	Position
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	case Eof:
		return "<eof>"
	case Name:
		prefix = "name"
	case Variable:
		prefix = "variable"
	case Current:
		return "<current>"
	case Parent:
		return "<parent>"
	case Attr:
		return "<attribute>"
	case Literal:
		prefix = "literal"
	case Number:
		prefix = "number"
	case Slash:
		return "<path>"
	case Deep:
		return "<deep-path>"
	case Wildcard:
		return "<wildcard>"
	case Pipe:
		return "<pipe>"
	case Comma:
		return "<comma>"
	case Namespace:
		return "<namespace>"
	case AxisSep:
		return "<axis>"
	case BegPred:
		return "<beg-predicate>"
	case EndPred:
		return "<end-predicate>"
	case BegGrp:
		return "<beg-group>"
	case EndGrp:
		return "<end-group>"
	case Eq:
		return "<equal>"
	case Ne:
		return "<notequal>"
	case Invalid:
		return "<invalid>"
	default:
		return "<unknown>"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}

const (
	Eof rune = iota
	Name
	Variable
	Current
	Parent
	Attr
	Literal
	Number
	Namespace
	Slash
	Deep
	Wildcard
	Pipe
	Comma
	AxisSep
	BegPred
	EndPred
	BegGrp
	EndGrp
	Eq
	Ne
	Invalid
)

type Scanner struct {
	input io.RuneScanner
	char  rune

	Position
	old Position

	eof bool
	err error
	str bytes.Buffer
}

func Scan(r io.Reader) *Scanner {
	scan := Scanner{
		input: bufio.NewReader(r),
	}
	scan.Line++
	scan.read()
	return &scan
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) Scan() Token {
	var tok Token
	if s.err != nil {
		tok.Type = Invalid
		return tok
	}
	s.skipBlanks()
	if s.done() {
		tok.Type = Eof
		return tok
	}
	defer s.str.Reset()

	tok.Position = s.Position
	switch {
	case isQuote(s.char):
		s.scanLiteral(&tok)
	case isNumber(s.char) || (s.char == '-' && isNumber(s.peek())) || (s.char == '.' && isNumber(s.peek())):
		s.scanNumber(&tok)
	case isLetter(s.char):
		s.scanName(&tok)
	case isVariable(s.char):
		s.scanVariable(&tok)
	case isAttribute(s.char):
		s.scanAttribute(&tok)
	case isDelim(s.char):
		s.scanDelimiter(&tok)
	default:
		s.read()
		tok.Type = Invalid
	}
	return tok
}

func (s *Scanner) scanLiteral(tok *Token) {
	quote := s.char
	s.read()
	for !s.done() && s.char != quote {
		s.write()
		s.read()
	}
	tok.Type = Literal
	tok.Literal = s.str.String()
	if s.char != quote {
		tok.Type = Invalid
	} else {
		s.read()
	}
}

func (s *Scanner) scanNumber(tok *Token) {
	tok.Type = Number
	if s.char == '-' {
		s.write()
		s.read()
	}
	if s.char == '0' && isNumber(s.peek()) {
		tok.Type = Invalid
	}
	for !s.done() && isNumber(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
	if s.char != '.' {
		return
	}
	s.write()
	s.read()
	for !s.done() && isNumber(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
}

func (s *Scanner) scanName(tok *Token) {
	for !s.done() && (isAlpha(s.char) || s.char == '-' || s.char == '.') {
		s.write()
		s.read()
	}
	tok.Type = Name
	tok.Literal = s.str.String()
}

func (s *Scanner) scanVariable(tok *Token) {
	s.read()
	tok.Type = Variable
	if !isLetter(s.char) {
		tok.Type = Invalid
	}
	for !s.done() && isAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.str.String()
}

func (s *Scanner) scanAttribute(tok *Token) {
	s.read()
	tok.Type = Attr
}

func (s *Scanner) scanDelimiter(tok *Token) {
	tok.Type = Invalid
	switch s.char {
	case '/':
		tok.Type = Slash
		if k := s.peek(); k == s.char {
			s.read()
			tok.Type = Deep
		}
	case ',':
		tok.Type = Comma
	case '|':
		tok.Type = Pipe
	case '(':
		tok.Type = BegGrp
	case ')':
		tok.Type = EndGrp
	case '[':
		tok.Type = BegPred
	case ']':
		tok.Type = EndPred
	case '.':
		tok.Type = Current
		if k := s.peek(); k == s.char {
			s.read()
			tok.Type = Parent
		}
	case '*':
		tok.Type = Wildcard
	case ':':
		tok.Type = Namespace
		if k := s.peek(); k == s.char {
			tok.Type = AxisSep
			s.read()
		}
	case '=':
		tok.Type = Eq
	case '!':
		if k := s.peek(); k != '=' {
			break
		}
		s.read()
		tok.Type = Ne
	default:
	}
	s.read()
}

func (s *Scanner) writeRune(c rune) {
	s.str.WriteRune(c)
}

func (s *Scanner) write() {
	s.writeRune(s.char)
}

func (s *Scanner) read() {
	if s.eof || s.err != nil {
		return
	}
	s.old = s.Position
	if s.char == '\n' {
		s.Line++
		s.Column = 0
	}
	s.Column++

	char, _, err := s.input.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			s.eof = true
			s.char = 0
		} else {
			s.err = err
		}
		return
	}
	s.char = char
}

func (s *Scanner) peek() rune {
	r, _, err := s.input.ReadRune()
	if err != nil {
		return 0
	}
	s.input.UnreadRune()
	return r
}

func (s *Scanner) done() bool {
	return s.eof || s.err != nil
}

func (s *Scanner) skipBlanks() {
	for isBlank(s.char) {
		s.read()
	}
}

func isNumber(c rune) bool {
	return c >= '0' && c <= '9'
}

func isLower(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isUpper(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLetter(c rune) bool {
	return isLower(c) || isUpper(c) || c == '_'
}

func isAlpha(c rune) bool {
	return isLetter(c) || isNumber(c)
}

func isQuote(c rune) bool {
	return c == '\'' || c == '"'
}

func isVariable(c rune) bool {
	return c == '$'
}

func isAttribute(c rune) bool {
	return c == '@'
}

func isDelim(c rune) bool {
	return c == '/' || c == ',' || c == '|' ||
		c == '(' || c == ')' || c == '[' || c == ']' ||
		c == '.' || c == '*' || c == ':' ||
		c == '=' || c == '!'
}

func isBlank(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
