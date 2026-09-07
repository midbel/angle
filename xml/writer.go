package xml

import (
	"bufio"
	"io"
	"strings"
)

type Writer struct {
	ws      *bufio.Writer
	depth   int
	compact bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		ws: bufio.NewWriter(w),
	}
}

func (w *Writer) SetCompact(c bool) {
	w.compact = c
}

func (w *Writer) Write(doc *Document) error {
	return w.ws.Flush()
}

type Formatter struct {
	ws      *bufio.Writer
	scan    *scanner
	depth   int
	compact bool
}

func NewFormatter(w io.Writer, r io.Reader) *Formatter {
	f := Formatter{
		ws:   bufio.NewWriter(w),
		scan: createScanner(r),
	}
	return &f
}

func (f *Formatter) SetCompact(c bool) {
	f.compact = c
}

func (f *Formatter) Format() error {
	for {
		tok := f.scan.Scan()
		if tok.Type == TokEof {
			break
		}
		var err error
		switch tok.Type {
		case TokOpenTag:
			err = f.writeOpenTag()
		case TokOpenPI:
			err = f.writePI()
		case TokComment:
			f.writeComment(tok)
		case TokCDATA:
			f.writeCharData(tok)
		case TokReference:
			f.ws.WriteString(tok.Literal)
		case TokText:
			f.writeText(tok)
		default:
		}
		if err != nil {
			return err
		}
	}
	return f.ws.Flush()
}

func (f *Formatter) writeOpenTag() error {
	tok := f.scan.Scan()
	switch tok.Type {
	case TokSlash:
		f.depth--
		return f.writeEndElement()
	case TokName:
		f.depth++
		return f.writeStartElement(tok)
	default:
		return ErrSyntax
	}
}

func (f *Formatter) writeStartElement(tok Token) error {
	f.ws.WriteRune(langle)
	f.ws.WriteString(tok.Literal)

	for {
		tok = f.scan.Scan()
		if tok.Type != TokName {
			break
		}
		f.ws.WriteRune(space)
		f.ws.WriteString(tok.Literal)
		if tok = f.scan.Scan(); tok.Type != TokEqual {
			return ErrAttribute
		}
		f.ws.WriteRune(equal)
		if tok = f.scan.Scan(); tok.Type != TokString {
			return ErrAttribute
		}
		f.writeString(tok)
	}
	if tok.Type == TokSlash {
		f.ws.WriteRune(slash)
		tok = f.scan.Scan()
	}
	if tok.Type != TokCloseTag {
		return ErrElement
	}
	f.ws.WriteRune(rangle)
	return nil
}

func (f *Formatter) writeEndElement() error {
	f.ws.WriteRune(langle)
	f.ws.WriteRune(slash)

	tok := f.scan.Scan()
	if tok.Type != TokName {
		return ErrElement
	}
	f.ws.WriteString(tok.Literal)
	if tok = f.scan.Scan(); tok.Type != TokCloseTag {
		return ErrElement
	}
	f.ws.WriteRune(rangle)
	return nil
}

func (f *Formatter) writePI() error {
	f.ws.WriteRune(langle)
	f.ws.WriteRune(question)

	tok := f.scan.Scan()
	if tok.Type != TokName {
		return ErrElement
	}
	f.ws.WriteString(tok.Literal)

	if tok = f.scan.Scan(); tok.Type == TokString {
		f.ws.WriteRune(space)
		f.ws.WriteString(tok.Literal)
	}
	if tok = f.scan.Scan(); tok.Type != TokClosePI {
		return ErrElement
	}
	f.ws.WriteRune(question)
	f.ws.WriteRune(rangle)
	return nil
}

func (f *Formatter) writeComment(tok Token) {
	f.ws.WriteRune(langle)
	f.ws.WriteRune(dash)
	f.ws.WriteRune(dash)
	f.ws.WriteString(tok.Literal)
	f.ws.WriteRune(dash)
	f.ws.WriteRune(dash)
	f.ws.WriteRune(rangle)
}

func (f *Formatter) writeText(tok Token) {
	lit := tok.Literal
	if f.compact {
		lit = strings.TrimSpace(lit)
		if lit == "" {
			return
		}
	}
	f.ws.WriteString(lit)
}

func (f *Formatter) writeCharData(tok Token) {
	f.ws.WriteRune(langle)
	f.ws.WriteRune(question)
	f.ws.WriteRune(lsquare)
	f.ws.WriteString("CDATA")
	f.ws.WriteRune(lsquare)
	f.ws.WriteString(tok.Literal)
	f.ws.WriteRune(rsquare)
	f.ws.WriteRune(rsquare)
	f.ws.WriteRune(rangle)
}

func (f *Formatter) writeString(tok Token) {
	f.ws.WriteRune(dquote)
	f.ws.WriteString(tok.Literal)
	f.ws.WriteRune(dquote)
}
