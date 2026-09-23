package svg

import (
	"errors"
	"io"
	"strings"

	"github.com/midbel/angle/xml"
)

var (
	ErrNegative = errors.New("negative value")
	ErrRange    = errors.New("value out of range")
)

type Element interface {
	Element() xml.Node
}

type Document struct {
	origin   Point
	size     Size
	children []Element
}

func NewDocument(width, height float64) *Document {
	return &Document{
		size:   NewSize(width, height),
		origin: NewPoint(0, 0),
	}
}

func (d *Document) Render(w io.Writer) error {
	e := xml.NewEncoder(w)
	x := xml.NewDocument(d.Element())
	return e.Encode(x)
}

func (d *Document) Append(el Element) {
	d.children = append(d.children, el)
}

func (d *Document) Element() xml.Node {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("viewBox"), d.viewBox()),
	}
	attrs = append(attrs, d.size.attributes()...)
	attrs = append(attrs, d.origin.attributes()...)

	ns := xml.Namespace{
		Prefix: "",
		URI:    "http://www.w3.org/2000/svg",
	}

	el := xml.NewElement(xml.NewName("svg"))
	el.Attributes = cleanAttrs(attrs)
	el.NS = append(el.NS, ns)

	for _, c := range d.children {
		el.Children = append(el.Children, c.Element())
	}
	return el
}

func (d *Document) Resize(width, height float64) *Document {
	d.size.Resize(width, height)
	return d
}

func (d *Document) Move(x, y float64) *Document {
	d.origin.Move(x, y)
	return d
}

func (d *Document) viewBox() string {
	var str strings.Builder
	str.WriteString(f2s(d.origin.x))
	str.WriteRune(' ')
	str.WriteString(f2s(d.origin.y))
	str.WriteRune(' ')
	str.WriteString(f2s(d.size.width))
	str.WriteRune(' ')
	str.WriteString(f2s(d.size.height))
	return str.String()
}

type Group struct {
	children []Element
}

func NewGroup() *Group {
	return &Group{}
}

func (g *Group) Element() xml.Node {
	el := xml.NewElement(xml.NewName("g"))
	for _, c := range g.children {
		el.Children = append(el.Children, c.Element())
	}
	return el
}
