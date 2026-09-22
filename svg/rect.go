package svg

import (
	"github.com/midbel/angle/xml"
)

type Rect struct {
	x      float64
	y      float64
	width  float64
	height float64

	rx float64
	ry float64

	fill        string
	stroke      string
	strokeWidth float64
}

func NewRect(x, y, width, height float64) *Rect {
	return &Rect{
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}
}

func (r *Rect) Element() xml.Node {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("x"), f2s(r.x)),
		xml.NewAttribute(xml.NewName("y"), f2s(r.y)),
		xml.NewAttribute(xml.NewName("width"), f2s(r.width)),
		xml.NewAttribute(xml.NewName("width"), f2s(r.height)),
		xml.NewAttribute(xml.NewName("fill"), r.fill),
		xml.NewAttribute(xml.NewName("stroke"), r.stroke),
	}
	if r.rx > 0 {
		a := xml.NewAttribute(xml.NewName("rx"), f2s(r.rx))
		attrs = append(attrs, a)
	}
	if r.ry > 0 {
		a := xml.NewAttribute(xml.NewName("ry"), f2s(r.ry))
		attrs = append(attrs, a)
	}
	if r.strokeWidth > 0 {
		a := xml.NewAttribute(xml.NewName("stroke-width"), f2s(r.strokeWidth))
		attrs = append(attrs, a)
	}
	el := xml.NewElement(xml.NewName("rect"))
	el.Attributes = cleanAttrs(attrs)
	return el
}

func (r *Rect) Resize(width, height float64) *Rect {
	r.width = width
	r.height = height
	return r
}

func (r *Rect) Move(x, y float64) *Rect {
	r.x = x
	r.y = y
	return r
}

func (r *Rect) Radius(rx, ry float64) *Rect {
	r.rx = rx
	r.ry = ry
	return r
}

func (r *Rect) Fill(fill string) *Rect {
	r.fill = fill
	return r
}

func (r *Rect) Stroke(stroke string) *Rect {
	r.stroke = stroke
	return r
}

func (r *Rect) StrokeWidth(width float64) *Rect {
	r.strokeWidth = width
	return r
}
