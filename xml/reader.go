package xml

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrSyntax      = errors.New("xml: syntax error")
	ErrName        = errors.New("xml: invalid name")
	ErrReference   = errors.New("xml: invalid reference")
	ErrElement     = errors.New("xml: invalid element")
	ErrAttribute   = errors.New("xml: invalid attribute")
	ErrNamespace   = errors.New("xml: invalid namespace")
	ErrDeclaration = errors.New("xml: invalid declaration")
	ErrDocument    = errors.New("xml: invalid document")
	ErrEOF         = errors.New("xml: unexpected end of input")
)

type Error struct {
	Err error
	Position
}

func (e Error) Error() string {
	return fmt.Sprintf("[%d:%d] %s", e.Line, e.Column, e.Err)
}

func (e Error) Unwrap() error {
	return e.Err
}

type Handler interface {
	OnStartElement(Element) error
	OnCloseElement(Name) error
	OnText(Text) error
	OnComment(Comment) error
	OnPI(PI) error
}

type Decoder struct {
	reader *Reader

	doc   *Document
	stack []*Element
}

func Decode(r io.Reader) (*Document, error) {
	b := NewDecoder(r)
	return b.Decode()
}

func NewDecoder(r io.Reader) *Decoder {
	b := Decoder{
		reader: NewReader(r),
		doc:    new(Document),
	}
	return &b
}

func (b *Decoder) Decode() (*Document, error) {
	return b.doc, b.reader.Read(b)
}

func (b *Decoder) OnStartElement(el Element) error {
	if n := len(b.stack); n == 0 {
		b.doc.Children = append(b.doc.Children, &el)
	} else {
		b.stack[n-1].Children = append(b.stack[n-1].Children, &el)
	}
	b.stack = append(b.stack, &el)
	return nil
}

func (b *Decoder) OnCloseElement(x Name) error {
	if n := len(b.stack); n == 0 {
		// TODO
	} else {
		b.stack = b.stack[:n-1]
	}
	return nil
}

func (b *Decoder) OnText(t Text) error {
	if n := len(b.stack); n == 0 {
		b.doc.Children = append(b.doc.Children, &t)
	} else {
		b.stack[n-1].Children = append(b.stack[n-1].Children, &t)
	}
	return nil
}

func (b *Decoder) OnComment(c Comment) error {
	if n := len(b.stack); n == 0 {
		b.doc.Children = append(b.doc.Children, &c)
	} else {
		b.stack[n-1].Children = append(b.stack[n-1].Children, &c)
	}
	return nil
}

func (b *Decoder) OnPI(p PI) error {
	if n := len(b.stack); n == 0 {
		b.doc.Children = append(b.doc.Children, &p)
	} else {
		b.stack[n-1].Children = append(b.stack[n-1].Children, &p)
	}
	return nil
}

type context struct {
	Name
	NS []Namespace
}

type Reader struct {
	scan *scanner
	curr Token
	peek Token

	hasRoot bool
	stack   []context
}

func NewReader(r io.Reader) *Reader {
	rs := &Reader{
		scan: createScanner(r),
	}
	rs.next()
	rs.next()
	return rs
}

func (r *Reader) Read(handler Handler) error {
	return r.startDocument(handler)
}

func (r *Reader) startDocument(handler Handler) error {
	for !r.done() {
		if err := r.checkDocumentWhitespace(); err != nil {
			return err
		}
		if err := r.startNode(handler); err != nil {
			return err
		}
	}
	if !r.hasRoot {
		return ErrDocument
	}
	if len(r.stack) > 0 {
		return r.createError(ErrEOF)
	}
	return nil
}

func (r *Reader) startNode(handler Handler) error {
	var err error
	switch {
	case r.is(TokOpenTag):
		err = r.readElement(handler)
	case r.is(TokOpenPI):
		err = r.readPI(handler)
	case r.is(TokComment):
		err = r.readComment(handler)
	case r.is(TokCDATA):
		err = r.readCharData(handler)
	case r.is(TokText) || r.is(TokReference):
		err = r.readText(handler)
	case r.is(TokEof):
		err = io.EOF
	default:
		err = r.syntaxError()
	}
	return err
}

func (r *Reader) readPI(handler Handler) error {
	r.next()
	name, err := r.decodeName()
	if err != nil {
		return err
	}
	pi := PI{
		Name: name,
	}
	if r.is(TokString) {
		pi.Data = r.currentLiteral()
		if name.Local == "xml" {
			// TODO
		}
		r.next()
	}
	if !r.is(TokClosePI) {
		return r.syntaxError()
	}
	r.next()
	return handler.OnPI(pi)
}

func (r *Reader) readElement(handler Handler) error {
	r.next()
	var err error
	switch {
	case r.is(TokName):
		err = r.readStartElement(handler)
	case r.is(TokSlash):
		err = r.readCloseElement(handler)
	default:
		return r.syntaxError()
	}
	return err
}

func (r *Reader) readStartElement(handler Handler) error {
	if err := r.checkRootElement(); err != nil {
		return err
	} else {
		defer r.setRoot()
	}
	name, err := r.decodeName()
	if err != nil {
		return err
	}
	el := Element{
		Name: name,
	}
	for !r.done() && !(r.is(TokSlash) || r.is(TokCloseTag)) {
		attr, err := r.decodeAttr()
		if err != nil {
			return err
		}
		if attr.Local == "xmlns" || attr.Prefix == "xmlns" {
			ns := Namespace{
				Prefix: attr.Local,
				URI:    attr.Value,
			}
			if attr.Local == "xmlns" {
				ns.Prefix = ""
			}
			el.NS = append(el.NS, ns)
			continue
		}
		el.Attributes = append(el.Attributes, attr)
	}
	var selfClosed bool
	if selfClosed = r.is(TokSlash); selfClosed {
		r.next()
		if !r.is(TokCloseTag) {
			return r.syntaxError()
		}
	}
	if !r.is(TokCloseTag) {
		return r.syntaxError()
	}
	ctx := context{
		Name: el.Name,
		NS:   el.NS,
	}
	r.pushContext(ctx)
	r.updateElementName(&el)

	ctx.Name = el.Name
	ctx.NS = el.NS
	r.replaceTop(ctx)

	if err := r.checkDuplicateAttributes(el); err != nil {
		return err
	}

	r.next()
	if err = handler.OnStartElement(el); err != nil {
		return err
	}
	if selfClosed {
		err = handler.OnCloseElement(el.Name)
		r.popContext()
	}
	return err
}

func (r *Reader) readCloseElement(handler Handler) error {
	r.next()
	name, err := r.decodeName()
	if err != nil {
		return err
	}
	ns, ok := r.resolveNamespace(name.Prefix)
	if ok {
		name.Namespace = ns
	} else {
		if name.Prefix != "" {
			return ErrNamespace
		}
	}
	if !r.is(TokCloseTag) {
		return r.syntaxError()
	}
	r.next()

	if ctx, ok := r.popContext(); !ok || !ctx.Equal(name) {
		return r.createError(ErrElement)
	}
	return handler.OnCloseElement(name)
}

func (r *Reader) readCharData(handler Handler) error {
	defer r.next()
	text := Text{
		Value: r.currentLiteral(),
	}
	return handler.OnText(text)
}

func (r *Reader) readText(handler Handler) error {
	var str strings.Builder
	for !r.done() && (r.is(TokText) || r.is(TokReference)) {
		switch {
		case r.is(TokReference):
			ref, err := r.decodeReference()
			if err != nil {
				return err
			}
			str.WriteString(ref)
		case r.is(TokText):
			str.WriteString(r.currentLiteral())
			r.next()
		default:
		}
	}
	txt := Text{
		Value: str.String(),
	}
	return handler.OnText(txt)
}

func (r *Reader) readComment(handler Handler) error {
	defer r.next()
	comment := Comment{
		Value: r.currentLiteral(),
	}
	return handler.OnComment(comment)
}

func (r *Reader) updateElementName(el *Element) error {
	ns, ok := r.resolveNamespace(el.Prefix)
	if ok {
		el.Namespace = ns
	} else {
		if el.Prefix != "" {
			return ErrNamespace
		}
	}
	for i, a := range el.Attributes {
		if a.Prefix == "" {
			continue
		}
		ns, ok := r.resolveNamespace(a.Name.Prefix)
		if ok {
			a.Name.Namespace = ns
			el.Attributes[i] = a
		} else {
			return ErrNamespace
		}
	}
	return nil
}

func (r *Reader) checkDuplicateAttributes(el Element) error {
	seen := make(map[string]struct{})
	for _, a := range el.Attributes {
		n := a.LexicalName()
		if _, ok := seen[n]; ok {
			return ErrAttribute
		}
		seen[n] = struct{}{}
	}
	return nil
}

func (r *Reader) decodeName() (Name, error) {
	if !r.is(TokName) {
		return Name{}, r.syntaxError()
	}
	defer r.next()
	return ParseName(r.currentLiteral())
}

func (r *Reader) decodeAttr() (Attribute, error) {
	name, err := r.decodeName()
	if err != nil {
		return Attribute{}, err
	}
	if !r.is(TokEqual) {
		return Attribute{}, r.syntaxError()
	}
	r.next()
	if !r.is(TokString) {
		return Attribute{}, r.syntaxError()
	}
	attr := Attribute{
		Name:  name,
		Value: r.currentLiteral(),
	}
	r.next()
	return attr, nil
}

func (r *Reader) decodeReference() (string, error) {
	defer r.next()

	str := r.currentLiteral()

	switch str {
	case "&lt;":
		return "<", nil
	case "&gt;":
		return ">", nil
	case "&amp;":
		return "&", nil
	case "&apos;":
		return "'", nil
	case "&quot;":
		return `"`, nil
	}

	var (
		num int64
		err error
	)
	switch {
	case strings.HasPrefix(str, "&#x"):
		num, err = strconv.ParseInt(str[3:len(str)-1], 16, 32)
	case strings.HasPrefix(str, "&#"):
		num, err = strconv.ParseInt(str[2:len(str)-1], 10, 32)
	default:
		return "", r.createError(ErrReference)
	}
	if err != nil {
		return "", r.createError(ErrReference)
	}
	char := rune(num)
	if !IsValidChar(char) {
		return "", r.createError(ErrReference)
	}
	return string(char), nil
}

func (r *Reader) is(kind Type) bool {
	return r.curr.Type == kind
}

func (r *Reader) currentLiteral() string {
	return r.curr.Literal
}

func (r *Reader) done() bool {
	return r.is(TokEof)
}

func (r *Reader) next() {
	r.curr = r.peek
	r.peek = r.scan.Scan()
}

func (r *Reader) createError(err error) error {
	return Error{
		Err:      err,
		Position: r.curr.Position,
	}
}

func (r *Reader) syntaxError() error {
	return r.createError(ErrSyntax)
}

func (r *Reader) resolveNamespace(prefix string) (Namespace, bool) {
	for i := len(r.stack) - 1; i >= 0; i-- {
		x := slices.IndexFunc(r.stack[i].NS, func(n Namespace) bool {
			return n.Prefix == prefix
		})
		if x >= 0 {
			return r.stack[i].NS[x], true
		}
	}
	return Namespace{}, false
}

func (r *Reader) pushContext(ctx context) {
	r.stack = append(r.stack, ctx)
}

func (r *Reader) replaceTop(ctx context) {
	if x := len(r.stack); x > 0 {
		r.stack[x-1] = ctx
	}
}

func (r *Reader) popContext() (context, bool) {
	var (
		ctx context
		ok  bool
	)
	if x := len(r.stack); x > 0 {
		ctx = r.stack[x-1]
		r.stack = r.stack[:x-1]
		ok = true
	}
	return ctx, ok
}

func (r *Reader) checkDocumentWhitespace() error {
	if !r.hasRoot {
		if r.is(TokCDATA) || r.is(TokReference) {
			return r.syntaxError()
		}
		if r.is(TokText) {
			str := strings.TrimSpace(r.currentLiteral())
			if len(str) != 0 {
				return r.syntaxError()
			}
		}
	}
	if r.hasRoot && len(r.stack) == 0 {
		if r.is(TokCDATA) || r.is(TokReference) {
			return r.syntaxError()
		}
		if r.is(TokText) {
			str := strings.TrimSpace(r.currentLiteral())
			if len(str) != 0 {
				return r.syntaxError()
			}
		}
	}
	return nil
}

func (r *Reader) setRoot() {
	r.hasRoot = true
}

func (r *Reader) checkRootElement() error {
	if r.hasRoot && len(r.stack) == 0 {
		return ErrDocument
	}
	r.hasRoot = true
	return nil
}
