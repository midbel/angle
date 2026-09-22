package svg

import (
	"io"

	"github.com/midbel/angle/xml"
)

type Element interface {
	Element() xml.Node
}

type Document struct {
	x        float64
	y        float64
	width    float64
	height   float64
	children []Element
}

func NewDocument(width, height float64) *Document {
	return &Document{
		width:  width,
		height: height,
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
	el := xml.NewElement(xml.NewName("svg"))
	el.Attributes = []xml.Attribute{
		xml.NewAttribute(xml.NewName("x"), f2s(d.x)),
		xml.NewAttribute(xml.NewName("y"), f2s(d.y)),
		xml.NewAttribute(xml.NewName("width"), f2s(d.width)),
		xml.NewAttribute(xml.NewName("height"), f2s(d.height)),
	}
	ns := xml.Namespace{
		Prefix: "",
		URI:    "http://www.w3.org/2000/svg",
	}
	el.NS = append(el.NS, ns)
	for _, c := range d.children {
		el.Children = append(el.Children, c.Element())
	}
	return el
}

func (d *Document) Resize(width, height float64) *Document {
	d.width = width
	d.height = height
	return d
}

func (d *Document) Move(x, y float64) *Document {
	d.x = x
	d.y = y
	return d
}
