package dispatch

import (
	"io"

	"github.com/midbel/angle/xml"
)

type (
	OnElementFunc func(xml.Element) error
	OnTextFunc    func(xml.Text) error
	OnPiFunc      func(xml.PI) error
)

type Reader struct {
	stack []listeners
	names []xml.Name
}

func New() *Reader {
	w := new(Reader)
	w.Push()
	return w
}

func (r *Reader) Read(rs io.Reader) error {
	rx := xml.NewReader(rs)
	return rx.Read(createHandler(r))
}

func (r *Reader) OnElement(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	r.OnOpen(name, fn)
	r.OnClose(name, fn)
}

func (r *Reader) OnOpen(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].onOpen[name] = fn
	}
}

func (r *Reader) OnAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	r.OnOpenAny(fn)
	r.OnCloseAny(fn)
}

func (r *Reader) OnOpenAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].onOpenAny = fn
	}
}

func (r *Reader) OnClose(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].onClose[name] = fn
	}
}

func (r *Reader) OnCloseAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].onCloseAny = fn
	}
}

func (r *Reader) OnPI(name xml.Name, fn OnPiFunc) {
	if fn == nil {
		return
	}
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].onPi[name] = fn
	}
}

func (r *Reader) OnText(fn OnTextFunc) {
	if fn == nil {
		return
	}
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].onText = fn
	}
}

func (r *Reader) Push() {
	s := listeners{
		onOpen:  make(map[xml.Name]OnElementFunc),
		onClose: make(map[xml.Name]OnElementFunc),
		onPi:    make(map[xml.Name]OnPiFunc),
	}
	r.stack = append(r.stack, s)
}

func (r *Reader) Pop() {
	if i := len(r.stack); i > 1 {
		r.stack = r.stack[:i-1]
	}
}

func (r *Reader) dispatchOpen(el xml.Element) error {
	r.pushName(el.Name)

	var found bool
	for i := len(r.stack) - 1; i >= 0; i-- {
		fn, ok := r.stack[i].onOpen[el.Name]
		if ok {
			found = true
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	if !found {
		return r.dispatchOpenAny(el)
	}
	return nil
}

func (r *Reader) dispatchOpenAny(el xml.Element) error {
	for i := len(r.stack) - 1; i >= 0; i-- {
		fn := r.stack[i].onOpenAny
		if fn != nil {
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (r *Reader) dispatchClose(el xml.Element) error {
	defer r.popName()

	var found bool
	for i := len(r.stack) - 1; i >= 0; i-- {
		fn, ok := r.stack[i].onClose[el.Name]
		if ok {
			found = true
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	if !found {
		return r.dispatchCloseAny(el)
	}
	return nil
}

func (r *Reader) dispatchCloseAny(el xml.Element) error {
	for i := len(r.stack) - 1; i >= 0; i-- {
		fn := r.stack[i].onCloseAny
		if fn != nil {
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (r *Reader) dispatchPI(pi xml.PI) error {
	for i := len(r.stack) - 1; i >= 0; i-- {
		fn, ok := r.stack[i].onPi[pi.Name]
		if ok {
			if err := fn(pi); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (r *Reader) dispatchText(txt xml.Text) error {
	for i := len(r.stack) - 1; i >= 0; i-- {
		fn := r.stack[i].onText
		if fn != nil {
			if err := fn(txt); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (r *Reader) pushName(qn xml.Name) {
	o.names = append(o.names, qn)
}

func (r *Reader) popName() {
	n := len(r.names)
	if n == 0 {
		return
	}
	r.names = r.names[:n-1]
}

type handler struct {
	in *Reader
}

func createHandler(in *Reader) xml.Handler {
	return &handler{
		in: in,
	}
}

func (h *handler) OnStartElement(el xml.Element, self bool) error {
	err := h.in.dispatchOpen(el)
	if err == nil && self {
		return h.in.dispathClose(el)
	}
	return err
}

func (h *handler) OnCloseElement(name xml.Name) error {
	el := xml.Element{
		Name: name,
	}
	return h.in.dispatchClose(el)
}

func (h *handler) OnText(txt xml.Text) error {
	return h.in.dispatchText(txt)
}

func (h *handler) OnComment(_ xml.Comment) error {
	// TODO
	return nil
}

func (h *handler) OnPI(pi xml.PI) error {
	return h.in.dispatchPI(pi)
}

type listeners struct {
	onOpen     map[xml.Name]OnElementFunc
	onClose    map[xml.Name]OnElementFunc
	onPi       map[xml.Name]OnPiFunc
	onOpenAny  OnElementFunc
	onCloseAny OnElementFunc
	onText     OnTextFunc
}