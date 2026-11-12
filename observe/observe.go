package observe

import (
	"io"

	"github.com/midbel/angle/xml"
)

type (
	OnElementFunc func(xml.Element) error
	OnTextFunc    func(xml.Text) error
	OnPiFunc      func(xml.PI) error
)

type Observer struct {
	stack []listeners
	names []xml.Name
}

func New() *Observer {
	w := new(Observer)
	w.Push()
	return w
}

func (o *Observer) Read(r io.Reader) error {
	rs := xml.NewReader(r)
	return rs.Read(createHandler(o))
}

func (o *Observer) OnElement(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	d.OnOpen(name, fn)
	d.OnClose(name, fn)
}

func (o *Observer) OnOpen(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(o.stack) - 1; i >= 0 {
		o.stack[i].onOpen[name] = fn
	}
}

func (o *Observer) OnAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	o.OnOpenAny(fn)
	o.OnCloseAny(fn)
}

func (o *Observer) OnOpenAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(o.stack) - 1; i >= 0 {
		o.stack[i].onOpenAny = fn
	}
}

func (o *Observer) OnClose(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(o.stack) - 1; i >= 0 {
		o.stack[i].onClose[name] = fn
	}
}

func (o *Observer) OnCloseAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	if i := len(o.stack) - 1; i >= 0 {
		o.stack[i].onCloseAny = fn
	}
}

func (o *Observer) OnPI(name xml.Name, fn OnPiFunc) {
	if fn == nil {
		return
	}
	if i := len(o.stack) - 1; i >= 0 {
		o.stack[i].onPi[name] = fn
	}
}

func (o *Observer) OnText(fn OnTextFunc) {
	if fn == nil {
		return
	}
	if i := len(o.stack) - 1; i >= 0 {
		o.stack[i].onText = fn
	}
}

func (o *Observer) Push() {
	s := listeners{
		onOpen:  make(map[xml.Name]OnElementFunc),
		onClose: make(map[xml.Name]OnElementFunc),
		onPi:    make(map[xml.Name]OnPiFunc),
	}
	o.stack = append(o.stack, s)
}

func (o *Observer) Pop() {
	if i := len(o.stack); i > 1 {
		o.stack = o.stack[:i-1]
	}
}

func (o *Observer) dispatchOpen(el xml.Element) error {
	o.pushName(el.Name)

	var found bool
	for i := len(o.stack) - 1; i >= 0; i-- {
		fn, ok := o.stack[i].onOpen[el.Name]
		if ok {
			found = true
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	if !found {
		return o.dispatchOpenAny(el)
	}
	return nil
}

func (o *Observer) dispatchOpenAny(el xml.Element) error {
	for i := len(o.stack) - 1; i >= 0; i-- {
		fn := o.stack[i].onOpenAny
		if fn != nil {
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (o *Observer) dispatchClose(el xml.Element) error {
	defer o.popName()

	var found bool
	for i := len(o.stack) - 1; i >= 0; i-- {
		fn, ok := o.stack[i].onClose[el.Name]
		if ok {
			found = true
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	if !found {
		return o.dispatchCloseAny(el)
	}
	return nil
}

func (o *Observer) dispatchCloseAny(el xml.Element) error {
	for i := len(o.stack) - 1; i >= 0; i-- {
		fn := o.stack[i].onCloseAny
		if fn != nil {
			if err := fn(el); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (o *Observer) dispatchPI(pi xml.PI) error {
	for i := len(o.stack) - 1; i >= 0; i-- {
		fn, ok := o.stack[i].onPi[pi.Name]
		if ok {
			if err := fn(pi); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (o *Observer) dispatchText(txt xml.Text) error {
	for i := len(o.stack) - 1; i >= 0; i-- {
		fn := o.stack[i].onText
		if fn != nil {
			if err := fn(txt); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func (o *Observer) pushName(qn xml.Name) {
	o.names = append(o.names, qn)
}

func (o *Observer) popName() {
	n := len(o.names)
	if n == 0 {
		return
	}
	o.names = o.names[:n-1]
}

type handler struct {
	obs *Observer
}

func createHandler(obs *Observer) xml.Handler {
	return &handler{
		obs: obs,
	}
}

func (h *handler) OnStartElement(el xml.Element, self bool) error {
	err := h.obs.dispatchOpen(el)
	if err == nil && self {
		return h.obs.dispathClose(el)
	}
	return err
}

func (h *handler) OnCloseElement(name xml.Name) error {
	el := xml.Element{
		Name: name,
	}
	return h.obs.dispatchClose(el)
}

func (h *handler) OnText(txt xml.Text) error {
	return h.obs.dispatchText(txt)
}

func (h *handler) OnComment(_ xml.Comment) error {
	// TODO
	return nil
}

func (h *handler) OnPI(pi xml.PI) error {
	return h.obs.dispatchPI(pi)
}

type listeners struct {
	onOpen     map[xml.Name]OnElementFunc
	onClose    map[xml.Name]OnElementFunc
	onPi       map[xml.Name]OnPiFunc
	onOpenAny  OnElementFunc
	onCloseAny OnElementFunc
	onText     OnTextFunc
}