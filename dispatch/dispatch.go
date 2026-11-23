package dispatch

import (
	"io"
	"slices"

	"github.com/midbel/angle/xml"
)

type Result uint16

const (
	Ok Result = 1 << iota
	Done

	Once Result = 1 << 8
)

func (r Result) Result() Result {
	return r &^ Once
}

func (r Result) Done() bool {
	return r.Result() == Done
}

func (r Result) Once() bool {
	return r&Once == Once
}

type (
	OnElementFunc func(*Reader, xml.Element) (Result, error)
	OnTextFunc    func(*Reader, xml.Text) (Result, error)
	OnPiFunc      func(*Reader, xml.PI) (Result, error)
)

type Reader struct {
	stack []*listeners
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

func (r *Reader) OnAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	r.OnOpenAny(fn)
	r.OnCloseAny(fn)
}

func (r *Reader) OnOpenOnce(name xml.Name, fn OnElementFunc) {
	cb := func(r *Reader, el xml.Element) (Result, error) {
		res, err := fn(r, el)
		return Once | res, err
	}
	r.OnOpen(name, cb)
}

func (r *Reader) OnOpen(name xml.Name, fn OnElementFunc) {
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].OnOpen(name, fn)
	}
}

func (r *Reader) OnOpenAnyOnce(fn OnElementFunc) {
	cb := func(r *Reader, el xml.Element) (Result, error) {
		res, err := fn(r, el)
		return Once | res, err
	}
	r.OnOpenAny(cb)
}

func (r *Reader) OnOpenAny(fn OnElementFunc) {
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].OnOpenAny(fn)
	}
}

func (r *Reader) OnCloseOnce(name xml.Name, fn OnElementFunc) {
	cb := func(r *Reader, el xml.Element) (Result, error) {
		res, err := fn(r, el)
		return Once | res, err
	}
	r.OnClose(name, cb)
}

func (r *Reader) OnClose(name xml.Name, fn OnElementFunc) {
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].OnClose(name, fn)
	}
}

func (r *Reader) OnCloseAnyOnce(fn OnElementFunc) {
	cb := func(r *Reader, el xml.Element) (Result, error) {
		res, err := fn(r, el)
		return Once | res, err
	}
	r.OnCloseAny(cb)
}

func (r *Reader) OnCloseAny(fn OnElementFunc) {
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].OnCloseAny(fn)
	}
}

func (r *Reader) OnPiOnce(name xml.Name, fn OnTextFunc) {
	cb := func(r *Reader, pi xml.PI) (Result, error) {
		res, err := fn(r, pi)
		return Once | res, err
	}
	r.OnPI(name, cb)
}

func (r *Reader) OnPI(name xml.Name, fn OnPiFunc) {
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].OnPi(name, fn)
	}
}

func (r *Reader) OnTextOnce(fn OnTextFunc) {
	cb := func(r *Reader, txt xml.Text) (Result, error) {
		res, err := fn(r, txt)
		return Once | res, err
	}
	r.OnText(cb)
}

func (r *Reader) OnText(fn OnTextFunc) {
	if i := len(r.stack) - 1; i >= 0 {
		r.stack[i].OnText(fn)
	}
}

func (r *Reader) Push() {
	s := &listeners{
		onOpen:  make(map[xml.Name][]OnElementFunc),
		onClose: make(map[xml.Name][]OnElementFunc),
		onPi:    make(map[xml.Name][]OnPiFunc),
	}
	r.stack = append(r.stack, s)
}

func (r *Reader) Pop() {
	if i := len(r.stack); i > 1 {
		r.stack = r.stack[:i-1]
	}
}

func (r *Reader) snapshot() []*listeners {
	return slices.Clone(r.stack)
}

func (r *Reader) dispatchOpen(el xml.Element) error {
	stack := r.snapshot()
	for i := len(stack) - 1; i >= 0; i-- {
		for _, fn := range stack[i].OpenFuncs(el.Name) {
			res, err := fn(r, el)
			if err != nil {
				return err
			}
			if res == Done {
				return nil
			}
		}
		for _, fn := range stack[i].OpenAnyFuncs() {
			res, err := fn(r, el)
			if err != nil {
				return err
			}
			if res == Done {
				return nil
			}
		}
	}
	return nil
}

func (r *Reader) dispatchOpenAny(el xml.Element) error {
	stack := r.snapshot()
	for i := len(stack) - 1; i >= 0; i-- {
		for _, fn := range stack[i].OpenAnyFuncs() {
			res, err := fn(r, el)
			if err != nil {
				return err
			}
			if res == Done {
				return nil
			}
		}
	}
	return nil
}

func (r *Reader) dispatchClose(el xml.Element) error {
	stack := r.snapshot()
	for i := len(stack) - 1; i >= 0; i-- {
		for _, fn := range stack[i].CloseFuncs(el.Name) {
			res, err := fn(r, el)
			if err != nil {
				return err
			}
			if res == Done {
				return nil
			}
		}
		for _, fn := range stack[i].CloseAnyFuncs() {
			res, err := fn(r, el)
			if err != nil {
				return err
			}
			if res == Done {
				return nil
			}
		}
	}
	return nil
}

func (r *Reader) dispatchCloseAny(el xml.Element) error {
	stack := r.snapshot()
	for i := len(stack) - 1; i >= 0; i-- {
		for _, fn := range stack[i].CloseAnyFuncs() {
			res, err := fn(r, el)
			if err != nil {
				return err
			}
			if res == Done {
				return nil
			}
		}
	}
	return nil
}

func (r *Reader) dispatchPI(pi xml.PI) error {
	stack := r.snapshot()
	for i := len(stack) - 1; i >= 0; i-- {
		for _, fn := range stack[i].PiFuncs(pi.Name) {
			res, err := fn(r, pi)
			if err != nil {
				return err
			}
			if res == Done {
				return nil
			}
		}
	}
	return nil
}

func (r *Reader) dispatchText(txt xml.Text) error {
	stack := r.snapshot()
	for i := len(stack) - 1; i >= 0; i-- {
		for j, fn := range stack[i].TextFuncs() {
			res, err := fn(r, txt)
			if res.Once() {
				stack[i].DetachText(j)
			}
			if err != nil {
				return err
			}
			if res.Done() {
				return nil
			}
		}
	}
	return nil
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
		return h.in.dispatchClose(el)
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
	onOpen     map[xml.Name][]OnElementFunc
	onClose    map[xml.Name][]OnElementFunc
	onPi       map[xml.Name][]OnPiFunc
	onOpenAny  []OnElementFunc
	onCloseAny []OnElementFunc
	onText     []OnTextFunc
}

func (i *listeners) TextFuncs() []OnTextFunc {
	return slices.Clone(i.onText)
}

func (i *listeners) OnText(fn OnTextFunc) {
	if fn == nil {
		return
	}
	i.onText = append(i.onText, fn)
}

func (i *listeners) DetachText(at int) {
	if at < 0 || at >= len(i.onText) {
		return
	}
	i.onText = slices.Delete(i.onText, at, at+1)
}

func (i *listeners) PiFuncs(name xml.Name) []OnPiFunc {
	return slices.Clone(i.onPi[name])
}

func (i *listeners) OnPi(name xml.Name, fn OnPiFunc) {
	if fn == nil {
		return
	}
	i.onPi[name] = append(i.onPi[name], fn)
}

func (i *listeners) DetachPI(name xml.PI, at int) {
	if at < 0 || at >= len(i.onPi[name]) {
		return
	}
	i.onPi[name] = slices.Delete(i.onPi[name], at, at+1)
}

func (i *listeners) OpenAnyFuncs() []OnElementFunc {
	return slices.Clone(i.onOpenAny)
}

func (i *listeners) OnOpenAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	i.onOpenAny = append(i.onOpenAny, fn)
}

func (i *listeners) DetachOpenAny(at int) {
	if at < 0 || at >= len(i.onOpenAny) {
		return
	}
	i.onOpenAny = slices.Delete(i.onOpenAny, at, at+1)
}

func (i *listeners) CloseAnyFuncs() []OnElementFunc {
	return slices.Clone(i.onCloseAny)
}

func (i *listeners) OnCloseAny(fn OnElementFunc) {
	if fn == nil {
		return
	}
	i.onCloseAny = append(i.onCloseAny, fn)
}

func (i *listeners) DetachCloseAny(at int) {
	if at < 0 || at >= len(i.onCloseAny) {
		return
	}
	i.onCloseAny = slices.Delete(i.onCloseAny, at, at+1)
}

func (i *listeners) OpenFuncs(name xml.Name) []OnElementFunc {
	return slices.Clone(i.onOpen[name])
}

func (i *listeners) OnOpen(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	i.onOpen[name] = append(i.onOpen[name], fn)
}

func (i *listeners) DetachOpen(name xml.Name, at int) {
	if at < 0 || at >= len(i.onOpen[name]) {
		return
	}
	i.onOpen[name] = slices.Delete(i.onOpen[name], at, at+1)	
}

func (i *listeners) CloseFuncs(name xml.Name) []OnElementFunc {
	return slices.Clone(i.onClose[name])
}

func (i *listeners) OnClose(name xml.Name, fn OnElementFunc) {
	if fn == nil {
		return
	}
	i.onClose[name] = append(i.onClose[name], fn)
}

func (i *listeners) DetachClose(name xml.Name, at int) {
	if at < 0 || at >= len(i.onClose[name]) {
		return
	}	
	i.onClose[name] = slices.Delete(i.onClose[name], at, at+1)
}
