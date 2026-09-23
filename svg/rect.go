package svg

import (
	"github.com/midbel/angle/xml"
)

type Rect struct {
	origin Point
	size   Size

	radius Radius

	fill   string
	stroke Stroke
}

func NewRect(x, y, width, height float64) *Rect {
	return &Rect{
		origin: NewPoint(x, y),
		size:   NewSize(width, height),
	}
}

func (r *Rect) Element() xml.Node {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("fill"), r.fill),
	}
	attrs = append(attrs, r.origin.attributes()...)
	attrs = append(attrs, r.size.attributes()...)
	attrs = append(attrs, r.radius.attributes()...)
	attrs = append(attrs, r.stroke.attributes()...)

	el := xml.NewElement(xml.NewName("rect"))
	el.Attributes = cleanAttrs(attrs)
	return el
}

func (r *Rect) Resize(width, height float64) *Rect {
	r.size.Resize(width, height)
	return r
}

func (r *Rect) Move(x, y float64) *Rect {
	r.origin.Move(x, y)
	return r
}

func (r *Rect) Radius(rx, ry float64) *Rect {
	r.radius = NewRadius(rx, ry)
	return r
}

func (r *Rect) Fill(fill string) *Rect {
	r.fill = fill
	return r
}

func (r *Rect) Stroke(stroke Stroke) *Rect {
	r.stroke = stroke
	return r
}

type Circle struct {
	origin Point
	radius float64

	fill   string
	stroke Stroke
}

func NewCircle(x, y, r float64) *Circle {
	return &Circle{
		origin: NewPoint(x, y),
		radius: r,
	}
}

func (c *Circle) Move(x, y float64) *Circle {
	c.origin.Move(x, y)
	return c
}

func (c *Circle) Resize(r float64) *Circle {
	c.radius = r
	return c
}

func (c *Circle) Fill(fill string) *Circle {
	c.fill = fill
	return c
}

func (c *Circle) Stroke(stroke Stroke) *Circle {
	c.stroke = stroke
	return c
}

type Ellipse struct {
	origin Point
	radius Radius

	fill   string
	stroke Stroke
}

func NewEllipse(x, y, rx, ry float64) *Ellipse {
	return &Ellipse{
		origin: NewPoint(x, y),
		radius: NewRadius(rx, ry),
	}
}

func (e *Ellipse) Move(x, y float64) *Ellipse {
	e.origin.Move(x, y)
	return e
}

func (e *Ellipse) Resize(rx, ry float64) *Ellipse {
	e.radius = NewRadius(rx, ry)
	return e
}

func (e *Ellipse) Fill(fill string) *Ellipse {
	e.fill = fill
	return e
}

func (e *Ellipse) Stroke(stroke Stroke) *Ellipse {
	e.stroke = stroke
	return e
}
