package xml

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

var ErrInput = errors.Input("bad data")

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
	TokPI
	TokComment
	TokCDATA
	TokDOCTYPE
	TokString
	TokEqual
	TokReference
	TokText
)

func (t Type) String() string {
	switch t {
	default:
		return "unknown"
	case TokOpenTag:
		return "open-tag"
	case TokCloseTag:
		return "close-tab"
	case TokSlash:
		return "slash"
	case TokName:
		return "name"
	case TokPI:
		return "pi"
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
	s.Position.Line++
	s.advance()
	return s, nil
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
	s.skipBlank()
	if s.done() {
		tok.Type = TokEof
		return tok
	}
	defer s.reset()

	tok.Position = s.Position
	switch {
	case s.char == langle:
	case s.char == rangle:
	case s.char == slash:
	case s.char == equal:
	case isQuote(s.char):
	default:
		tok.Type = TokInvalid
		s.write()
	}
	if tok.Type == TokInvalid {
		s.err = ErrInput
		tok.Literal = s.literal()
	}
	return tok
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
	s.char = c
	if s.char == cr && s.peek() == nl {
		s.char, _, _ = s.input.ReadRune()
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
