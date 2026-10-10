package svg

import (
	"io"

	"github.com/midbel/angle/xml"
)

type Element interface {
	Element() (xml.Node, error)
	Validate() error
}

type Document struct {
	*Attributes

	origin   Point
	size     Size
	box      ViewBox
	children []Element
}

func NewDocument(width, height float64) *Document {
	return &Document{
		Attributes: NewAttributes(),
		size:       NewSize(width, height),
		origin:     NewPoint(0, 0),
	}
}

func (d *Document) Render(w io.Writer) error {
	doc, err := d.Element()
	if err != nil {
		return err
	}
	e := xml.NewEncoder(w)
	return e.Encode(xml.NewDocument(doc))
}

func (d *Document) Add(el Element) {
	d.children = append(d.children, el)
}

func (d *Document) Element() (xml.Node, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	var attrs []xml.Attribute

	attrs = append(attrs, d.box.attributes()...)
	attrs = append(attrs, d.size.attributes()...)
	attrs = append(attrs, d.origin.attributes()...)
	attrs = append(attrs, d.Attributes.attributes()...)

	ns := xml.Namespace{
		Prefix: "",
		URI:    "http://www.w3.org/2000/svg",
	}

	el := xml.NewElement(xml.NewName("svg"))
	el.Attributes = cleanAttrs(attrs)
	el.NS = append(el.NS, ns)

	for _, c := range d.children {
		sub, err := c.Element()
		if err != nil {
			return nil, err
		}
		el.Children = append(el.Children, sub)
	}
	return el, nil
}

func (d *Document) Validate() error {
	if err := d.origin.Validate(); err != nil {
		return err
	}
	if err := d.size.Validate(); err != nil {
		return err
	}
	for _, e := range d.children {
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Document) Resize(width, height float64) *Document {
	d.size.Resize(width, height)
	return d
}

func (d *Document) Move(x, y float64) *Document {
	d.origin.Move(x, y)
	return d
}

type Group struct {
	*Attributes
	children []Element
}

func NewGroup() *Group {
	return &Group{
		Attributes: NewAttributes(),
	}
}

func (g *Group) Add(el Element) {
	g.children = append(g.children, el)
}

func (g *Group) Element() (xml.Node, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}
	el := xml.NewElement(xml.NewName("g"))
	el.Attributes = g.Attributes.attributes()
	for _, c := range g.children {
		sub, err := c.Element()
		if err != nil {
			return nil, err
		}
		el.Children = append(el.Children, sub)
	}
	return el, nil
}

func (g *Group) Validate() error {
	for _, e := range g.children {
		if err := e.Validate(); err != nil {
			return err
		}
	}
	return nil
}
