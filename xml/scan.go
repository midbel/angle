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
		return "close-tab"
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
	s.scan(&tok)
	return tok
}

func (s *scanner) scanDefault(tok *Token) {
	switch s.char {
	case langle:
		s.scanOpen(tok)
	default:
		tok.Type = TokInvalid
	}
}

func (s *scanner) scanOpen(tok *Token) {
	s.advance()
	switch s.char {
	case question:
		tok.Type = TokOpenPI
		s.scan = s.scanPIName
	default:
		tok.Type = TokInvalid
	}
	if tok.Type != TokInvalid {
		s.advance()
	}
}

func (s *scanner) scanPIName(tok *Token) {
	for !s.done() && !isBlank(s.char) && s.char != question {
		s.write()
		s.advance()
	}
	tok.Type = TokName
	tok.Literal = s.literal()
	if s.done() {
		tok.Type = TokInvalid
	}
	s.scan = s.scanPIData
}

func (s *scanner) scanPIData(tok *Token) {
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

const (
	langle    = '<'
	rangle    = '>'
	slash     = '/'
	dquote    = '"'
	squote    = '\''
	equal     = '='
	bang      = '!'
	minus     = '-'
	space     = ' '
	tab       = '\t'
	nl        = '\n'
	cr        = '\r'
	question  = '?'
	lsquare   = '['
	rsquare   = ']'
	ampersand = '&'
	semicolon = ';'
)

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
