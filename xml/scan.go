package xml

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"iter"
	"unicode"
)

var ErrInput = errors.New("bad data")

type Position struct {
	Line   int
	Column int
	Offset int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Literal string
	Type    Type
	Position
}

type Type rune

const (
	TokOpenTag Type = iota
	TokCloseTag
	TokSlash
	TokName
	TokOpenPI
	TokClosePI
	TokComment
	TokCDATA
	TokDOCTYPE
	TokString
	TokEqual
	TokReference
	TokText
	TokInvalid
	TokEof
)

func (t Type) String() string {
	switch t {
	default:
		return "unknown"
	case TokInvalid:
		return "invalid"
	case TokEof:
		return "eof"
	case TokOpenTag:
		return "open-tag"
	case TokCloseTag:
		return "close-tag"
	case TokSlash:
		return "slash"
	case TokName:
		return "name"
	case TokOpenPI:
		return "open-pi"
	case TokClosePI:
		return "close-pi"
	case TokComment:
		return "comment"
	case TokCDATA:
		return "cdata"
	case TokDOCTYPE:
		return "doctype"
	case TokString:
		return "string"
	case TokEqual:
		return "equal"
	case TokReference:
		return "reference"
	case TokText:
		return "text"
	}
}

type scanner struct {
	input *bufio.Reader
	err   error
	char  rune
	scan  func(*Token)

	buf *bytes.Buffer
	Position
}

func Scan(r io.Reader) iter.Seq2[Token, error] {
	it := func(yield func(Token, error) bool) {
		scan := createScanner(r)
		for !scan.done() {
			tok := scan.Scan()
			if !yield(tok, scan.Err()) {
				break
			}
		}
	}
	return it
}

func createScanner(r io.Reader) *scanner {
	input := bufio.NewReader(r)
	s := &scanner{
		input: input,
		buf:   new(bytes.Buffer),
	}
	s.scan = s.scanDefault
	s.Position.Line++
	s.advance()
	return s
}

func (s *scanner) Err() error {
	return s.err
}

func (s *scanner) Scan() Token {
	var tok Token
	if s.err != nil && !s.done() {
		tok.Type = TokInvalid
		return tok
	}
	if s.done() {
		tok.Type = TokEof
		return tok
	}
	defer s.reset()
	tok.Position = s.Position
	s.scan(&tok)
	return tok
}

func (s *scanner) scanDefault(tok *Token) {
	switch s.char {
	case langle:
		s.scanOpen(tok)
	default:
		s.scanText(tok)
	}
}

func (s *scanner) scanText(tok *Token) {
	for !s.done() && s.char != langle {
		s.write()
		s.advance()
	}
	tok.Type = TokText
	tok.Literal = s.literal()
}

func (s *scanner) scanOpen(tok *Token) {
	if k := s.peek(); k == question {
		tok.Type = TokOpenPI
		s.scan = s.scanPIName
		s.advance()
		// } else if k == slash {
		// 	tok.Type = TokSlash
		// 	s.scan = s.scanAfterSlash
		// 	s.advance()
	} else {
		tok.Type = TokOpenTag
		s.scan = s.scanTag
	}
	s.advance()
}

func (s *scanner) scanPIName(tok *Token) {
	s.scanName(tok)
	s.scan = s.scanPIData
}

func (s *scanner) scanPIData(tok *Token) {
	s.skipBlank()
	for !s.done() && s.char != question && s.peek() != rangle {
		s.write()
		s.advance()
	}
	tok.Type = TokString
	tok.Literal = s.literal()
	if s.done() {
		tok.Type = TokInvalid
	}
	s.scan = s.scanClosePI
}

func (s *scanner) scanClosePI(tok *Token) {
	tok.Type = TokInvalid
	if s.char == question && s.peek() == rangle {
		tok.Type = TokClosePI
		s.advance()
		s.advance()
	}
	s.scan = s.scanDefault
}

func (s *scanner) scanAfterSlash(tok *Token) {
	switch {
	case s.char == rangle:
		tok.Type = TokCloseTag
		s.scan = s.scanDefault
		s.advance()
	case isNameStart(s.char):
		s.scanName(tok)
		s.scan = s.scanTag
	default:
		tok.Type = TokInvalid
	}
}

func (s *scanner) scanTag(tok *Token) {
	s.skipBlank()
	switch {
	default:
	case isNameStart(s.char):
		s.scanName(tok)
	case s.char == slash:
		tok.Type = TokSlash
		s.advance()
		s.scan = s.scanAfterSlash
	case s.char == rangle:
		tok.Type = TokCloseTag
		s.scan = s.scanDefault
		s.advance()
	case s.char == equal:
		tok.Type = TokEqual
		s.scan = s.scanString
		s.advance()
	}
}

func (s *scanner) scanName(tok *Token) {
	if !isNameStart(s.char) {
		tok.Type = TokInvalid
		return
	}
	s.write()
	s.advance()
	for !s.done() && isNameChar(s.char) {
		s.write()
		s.advance()
	}
	tok.Type = TokName
	tok.Literal = s.literal()
	s.scan = s.scanTag
}

func (s *scanner) scanString(tok *Token) {
	if !isQuote(s.char) {
		tok.Type = TokInvalid
		return
	}
	opening := s.char
	s.advance()
	for !s.done() && s.char != opening {
		s.write()
		s.advance()
	}
	tok.Type = TokString
	tok.Literal = s.literal()
	if s.char != opening {
		tok.Type = TokInvalid
	} else {
		s.advance()
		s.scan = s.scanTag
	}
}

func (s *scanner) advance() {
	if s.err != nil {
		return
	}
	c, _, err := s.input.ReadRune()
	if err != nil {
		s.err = err
		s.char = 0
		return
	}
	s.Offset++
	s.char = c
	if s.char == cr && s.peek() == nl {
		s.char, _, _ = s.input.ReadRune()
		s.Offset++
	}

	if isNL(s.char) {
		s.Line += 1
		s.Column = 0
	}
	s.Column++
}

func (s *scanner) peek() rune {
	c, _, err := s.input.ReadRune()
	if err != nil {
		return 0
	}
	s.input.UnreadRune()
	return c
}

func (s *scanner) done() bool {
	return errors.Is(s.err, io.EOF) || s.char == 0
}

func (s *scanner) writeRune(r rune) {
	s.buf.WriteRune(r)
}

func (s *scanner) write() {
	s.writeRune(s.char)
}

func (s *scanner) reset() {
	s.buf.Reset()
}

func (s *scanner) literal() string {
	return s.buf.String()
}

func (s *scanner) skipSpace() {
	for isSpace(s.char) {
		s.advance()
	}
}

func (s *scanner) skipBlank() {
	for isBlank(s.char) {
		s.advance()
	}
}

const (
	langle     = '<'
	rangle     = '>'
	slash      = '/'
	dquote     = '"'
	squote     = '\''
	equal      = '='
	bang       = '!'
	minus      = '-'
	space      = ' '
	tab        = '\t'
	nl         = '\n'
	cr         = '\r'
	question   = '?'
	lsquare    = '['
	rsquare    = ']'
	ampersand  = '&'
	semicolon  = ';'
	colon      = ':'
	underscore = '_'
	dot        = '.'
)

func isNameStart(r rune) bool {
	return isLetter(r) || r == colon || r == underscore
}

func isNameChar(r rune) bool {
	return isLetter(r) || isDigit(r) ||
		r == dot || r == colon || r == underscore || r == minus
}

func isDigit(r rune) bool {
	return unicode.IsDigit(r)
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}

func isOpen(r rune) bool {
	return r == langle
}

func isClose(r rune) bool {
	return r == rangle
}

func isQuote(r rune) bool {
	return r == dquote || r == squote
}

func isAlpha(r rune) bool {
	return isLetter(r) || unicode.IsDigit(r)
}

func isBlank(r rune) bool {
	return isSpace(r) || isNL(r)
}

func isSpace(r rune) bool {
	return r == space || r == tab
}

func isNL(r rune) bool {
	return r == nl || r == cr
}
