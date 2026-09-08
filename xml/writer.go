package xml

import (
	"bufio"
	"fmt"
	"io"
)

type Encoder struct {
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{}
}

func (e *Encoder) Encode(doc *Document) error {
	return nil
}

type Formatter struct {
	ws *Writer
	rs *Reader
}

func NewFormatter(w io.Writer, r io.Reader) *Formatter {
	f := Formatter{
		ws: NewWriter(w),
		rs: NewReader(r),
	}
	return &f
}

func (f *Formatter) Format() error {
	if err := f.rs.Read(f); err != nil {
		return err
	}
	return f.ws.Flush()
}

func (f *Formatter) OnStartElement(el Element) error {
	return f.ws.StartElement(el.Name, el.Attributes, el.NS)
}

func (f *Formatter) OnCloseElement(n Name) error {
	return f.ws.CloseElement(n)
}

func (f *Formatter) OnText(t Text) error {
	return f.ws.Text(t.Value)
}

func (f *Formatter) OnComment(c Comment) error {
	return f.ws.Comment(c.Value)
}

func (f *Formatter) OnPI(p PI) error {
	return f.ws.PI(p.Name, p.Data)
}

type Writer struct {
	ws *bufio.Writer

	compact bool
	stack   []Name

	lastErr error
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		ws: bufio.NewWriter(w),
	}
}

func (w *Writer) SetCompact(compact bool) {
	w.compact = compact
}

func (w *Writer) Flush() error {
	if err := w.err(); err != nil {
		return err
	}
	return w.ws.Flush()
}

func (w *Writer) Empty(name Name, attrs []Attribute, ns []Namespace) error {
	if err := w.err(); err != nil {
		return err
	}
	w.writeRune(langle)
	if err := w.writeName(name); err != nil {
		return err
	}
	for _, n := range ns {
		w.writeRune(space)
		if err := w.writeNamespace(n); err != nil {
			return err
		}
	}
	for _, a := range attrs {
		w.writeRune(space)
		if err := w.writeAttribute(a); err != nil {
			return err
		}
	}
	w.writeRune(slash)
	w.writeRune(rangle)
	return w.err()
}

func (w *Writer) StartElement(name Name, attrs []Attribute, ns []Namespace) error {
	if err := w.err(); err != nil {
		return err
	}
	w.push(name)
	w.writeRune(langle)
	if err := w.writeName(name); err != nil {
		return err
	}
	for _, n := range ns {
		w.writeRune(space)
		if err := w.writeNamespace(n); err != nil {
			return err
		}
	}
	for _, a := range attrs {
		w.writeRune(space)
		if err := w.writeAttribute(a); err != nil {
			return err
		}
	}
	w.writeRune(rangle)
	return w.err()
}

func (w *Writer) CloseElement(name Name) error {
	if err := w.err(); err != nil {
		return err
	}

	last, ok := w.pop()
	if !ok || !last.Equal(name) {
		return ErrElement
	}

	w.writeRune(langle)
	w.writeRune(slash)
	if err := w.writeName(name); err != nil {
		return err
	}
	w.writeRune(rangle)
	return w.err()
}

func (w *Writer) Text(text string) error {
	if err := w.err(); err != nil {
		return err
	}
	return w.writeString(text)
}

func (w *Writer) Comment(comment string) error {
	return nil
	if err := w.err(); err != nil {
		return err
	}
	w.writeRune(langle)
	w.writeRune(dash)
	w.writeRune(dash)
	if err := w.writeString(comment); err != nil {
		return err
	}
	w.writeRune(dash)
	w.writeRune(dash)
	w.writeRune(rangle)
	return w.err()
}

func (w *Writer) PI(name Name, data string) error {
	if err := w.err(); err != nil {
		return err
	}
	w.writeRune(langle)
	w.writeRune(question)
	if err := w.writeName(name); err != nil {
		return err
	}
	w.writeRune(space)
	if err := w.writeString(data); err != nil {
		return err
	}
	w.writeRune(question)
	w.writeRune(rangle)
	return w.err()
}

func (w *Writer) writeNamespace(ns Namespace) error {
	if err := w.err(); err != nil {
		return err
	}
	if err := w.writeString("xmlns"); err != nil {
		return err
	}
	if ns.Prefix != "" {
		w.writeRune(colon)
		if err := w.writeString(ns.Prefix); err != nil {
			return err
		}
	}
	w.writeRune(equal)
	w.writeRune(dquote)
	if err := w.writeString(ns.URI); err != nil {
		return err
	}
	w.writeRune(dquote)
	return w.err()
}

func (w *Writer) writeAttribute(attr Attribute) error {
	if err := w.err(); err != nil {
		return err
	}
	if err := w.writeName(attr.Name); err != nil {
		return err
	}
	w.writeRune(equal)
	w.writeRune(dquote)
	if err := w.writeString(attr.Value); err != nil {
		return err
	}
	w.writeRune(dquote)
	return w.err()
}

func (w *Writer) writeName(name Name) error {
	if err := w.err(); err != nil {
		return err
	}
	if name.Prefix != "" {
		err := w.writeString(name.Prefix)
		if err != nil {
			return err
		}
		w.writeRune(colon)
	}
	if err := w.writeString(name.Local); err != nil {
		return err
	}
	return w.err()
}

func (w *Writer) writeString(str string) error {
	if err := w.err(); err != nil {
		return err
	}
	_, w.lastErr = w.ws.WriteString(str)
	return w.lastErr
}

func (w *Writer) writeRune(char rune) {
	if err := w.err(); err != nil {
		return
	}
	_, w.lastErr = w.ws.WriteRune(char)
}

func (w *Writer) nl() error {
	if err := w.err(); err != nil {
		return err
	}
	w.writeRune(nl)
	return w.err()
}

func (w *Writer) err() error {
	return w.lastErr
}

func (w *Writer) push(name Name) {
	w.stack = append(w.stack, name)
}

func (w *Writer) pop() (Name, bool) {
	var (
		last Name
		ok   bool
	)
	if x := len(w.stack); x > 0 {
		ok = true
		last = w.stack[x-1]
		w.stack = w.stack[:x-1]
	}
	return last, ok
}
