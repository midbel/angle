package xml

import (
	"bufio"
	"io"
	"strings"
)

type Encoder struct {
	writer *Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer: NewWriter(w),
	}
}

func (e *Encoder) SetCompact(compact bool) {
	e.writer.SetCompact(compact)
}

func (e *Encoder) Encode(doc *Document) error {
	for _, n := range doc.Children {
		if err := e.encodeNode(n); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeNode(n Node) error {
	var err error
	switch n := n.(type) {
	case *Element:
		err = e.encodeElement(n)
	case *PI:
		err = e.encodePI(n)
	case *Text:
		err = e.encodeText(n)
	case *Comment:
		err = e.encodeComment(n)
	default:
		err = ErrElement
	}
	return err
}

func (e *Encoder) encodeElement(el *Element) error {
	if err := e.writer.StartElement(el.Name, el.Attributes, el.NS); err != nil {
		return err
	}
	for _, n := range el.Children {
		if err := e.encodeNode(n); err != nil {
			return err
		}
	}
	return e.writer.CloseElement(el.Name)
}

func (e *Encoder) encodePI(pi *PI) error {
	return e.writer.PI(pi.Name, pi.Data)
}

func (e *Encoder) encodeText(txt *Text) error {
	return e.writer.Text(txt.Value)
}

func (e *Encoder) encodeComment(cmt *Comment) error {
	return e.writer.Comment(cmt.Value)
}

type Formatter struct {
	writer *Writer

	depth int

	offset int
	stack  []*layoutItem
	types  []itemType
}

func NewFormatter(w io.Writer) *Formatter {
	f := Formatter{
		writer: NewWriter(w),
	}
	return &f
}

func (f *Formatter) SetCompact(compact bool) {
	f.writer.SetCompact(compact)
}

func (f *Formatter) Format(r io.ReadSeeker) error {
	defer f.reset()
	root, err := f.analyze(r)
	if err != nil {
		return err
	}
	f.stack = flatten(root)
	f.stack = f.stack[1:]
	return f.format(r)
}

func (f *Formatter) OnStartElement(e Element) error {
	if f.offset >= len(f.stack) {
		return ErrSyntax
	}
	item := f.stack[f.offset]
	f.offset++
	if f.isBlock() {
		f.writer.NL()
		f.writer.Indent(f.depth)
	}
	err := f.writer.StartElement(e.Name, e.Attributes, e.NS)
	if err != nil {
		return err
	}
	f.types = append(f.types, item.typeOf())
	f.depth++
	return nil
}

func (f *Formatter) OnCloseElement(n Name) error {
	if f.isBlock() {
		f.writer.NL()
		f.writer.Indent(f.depth - 1)
	}
	if err := f.writer.CloseElement(n); err != nil {
		return err
	}
	if n := len(f.types); n > 0 {
		f.types = f.types[:n-1]
	}
	f.depth--
	return nil
}

func (f *Formatter) OnText(t Text) error {
	str := strings.TrimSpace(t.Value)
	if str == "" {
		return nil
	}
	return f.writer.Text(t.Value)
}

func (f *Formatter) OnComment(c Comment) error {
	if f.isBlock() && f.offset > 0 {
		if err := f.writer.NL(); err != nil {
			return err
		}
		if err := f.writer.Indent(f.depth); err != nil {
			return err
		}
	}
	if err := f.writer.Comment(c.Value); err != nil {
		return err
	}
	return nil
}

func (f *Formatter) OnPI(p PI) error {
	if f.isBlock() && f.offset > 0 {
		if err := f.writer.NL(); err != nil {
			return err
		}
		if err := f.writer.Indent(f.depth); err != nil {
			return err
		}
	}
	if err := f.writer.PI(p.Name, p.Data); err != nil {
		return err
	}
	return nil
}

func (f *Formatter) analyze(r io.ReadSeeker) (*layoutItem, error) {
	var (
		lh = newLayoutHandler()
		rs = NewReader(r)
	)
	if err := rs.Read(lh); err != nil {
		return nil, err
	}
	_, err := r.Seek(0, io.SeekStart)
	return lh.doc, err
}

func (f *Formatter) format(r io.Reader) error {
	rs := NewReader(r)
	if err := rs.Read(f); err != nil {
		return err
	}
	return f.writer.Flush()
}

func (f *Formatter) reset() {
	f.depth = 0
	f.offset = 0
	f.types = f.types[:0]
	f.stack = f.stack[:0]
}

func (f *Formatter) isBlock() bool {
	n := len(f.types)
	if n == 0 {
		return true
	}
	return f.types[n-1].Block()
}

type itemType uint8

const (
	typeBlock itemType = iota
	typeInline
	typeMixed
)

func (i itemType) Block() bool {
	return i == typeBlock
}

func (i itemType) String() string {
	switch i {
	case typeBlock:
		return "block"
	case typeInline:
		return "inline"
	case typeMixed:
		return "mixed"
	default:
		return "unknown"
	}
}

type layoutItem struct {
	Name
	Types    []NodeType
	Children []*layoutItem
}

func flatten(it *layoutItem) []*layoutItem {
	var list []*layoutItem
	list = append(list, it)
	if len(it.Children) == 0 {
		return list
	}
	for _, c := range it.Children {
		tmp := flatten(c)
		list = append(list, tmp...)
	}
	return list
}

func (i *layoutItem) root() bool {
	return i.Name.Zero()
}

func (i *layoutItem) typeOf() itemType {
	if len(i.Types) == 0 {
		return typeBlock
	}
	var (
		inline bool
		block  bool
	)
	for _, t := range i.Types {
		switch t {
		case ElementNode, CommentNode, PINode:
			block = true
		case TextNode:
			inline = true
		default:
		}
	}
	switch {
	case block && !inline:
		return typeBlock
	case inline && !block:
		return typeInline
	default:
		return typeMixed
	}
}

type layoutHandler struct {
	doc   *layoutItem
	stack []*layoutItem
}

func newLayoutHandler() *layoutHandler {
	h := layoutHandler{
		doc: new(layoutItem),
	}
	return &h
}

func (h *layoutHandler) OnStartElement(el Element) error {
	n := &layoutItem{
		Name: el.Name,
	}
	h.appendType(ElementNode)
	h.appendNode(n)
	h.stack = append(h.stack, n)
	return nil
}

func (h *layoutHandler) OnCloseElement(n Name) error {
	if n := len(h.stack); n == 0 {
		// TODO
	} else {
		h.stack = h.stack[:n-1]
	}
	return nil
}

func (h *layoutHandler) OnText(t Text) error {
	if s := strings.TrimSpace(t.Value); s != "" {
		h.appendType(TextNode)
	}
	return nil
}

func (h *layoutHandler) OnComment(c Comment) error {
	h.appendType(CommentNode)
	return nil
}

func (h *layoutHandler) OnPI(p PI) error {
	h.appendType(PINode)
	return nil
}

func (h *layoutHandler) appendType(kind NodeType) {
	if n := len(h.stack); n == 0 {
		h.doc.Types = append(h.doc.Types, kind)
	} else {
		h.stack[n-1].Types = append(h.stack[n-1].Types, kind)
	}
}

func (h *layoutHandler) appendNode(node *layoutItem) {
	if n := len(h.stack); n == 0 {
		h.doc.Children = append(h.doc.Children, node)
	} else {
		h.stack[n-1].Children = append(h.stack[n-1].Children, node)
	}
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

func (w *Writer) NL() error {
	return w.nl()
}

func (w *Writer) Indent(level int) error {
	if w.compact {
		return nil
	}
	if err := w.err(); err != nil {
		return err
	}
	for range level * 2 {
		w.writeRune(space)
	}
	return w.err()
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
	if w.compact {
		return nil
	}
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
