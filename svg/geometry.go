package svg

import (
	"strings"

	"github.com/midbel/angle/xml"
)

type Radius struct {
	x float64
	y float64
}

func NewRadius(x, y float64) Radius {
	return Radius{
		x: x,
		y: y,
	}
}

func (r *Radius) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("rx"), f2s(r.x)),
		xml.NewAttribute(xml.NewName("ry"), f2s(r.y)),
	}
}

func (r *Radius) Validate() error {
	if err := isFinite(r.x); err != nil {
		return err
	}
	if err := isFinite(r.y); err != nil {
		return err
	}
	if r.x < 0 || r.y < 0 {
		return negative("radius")
	}
	return nil
}

type ViewBox struct {
	origin Point
	size   Size
}

func NewViewBox(x, y, width, height float64) ViewBox {
	return ViewBox{
		origin: NewPoint(x, y),
		size:   NewSize(width, height),
	}
}

func (v ViewBox) Resize(width, height float64) ViewBox {
	v.size.Resize(width, height)
	return v
}

func (v ViewBox) Move(x, y float64) ViewBox {
	v.origin.Move(x, y)
	return v
}

func (v ViewBox) Validate() error {
	if err := v.origin.Validate(); err != nil {
		return err
	}
	if err := v.size.Validate(); err != nil {
		return err
	}
	return nil
}

func (v ViewBox) attributes() []xml.Attribute {
	if v.origin.Zero() && v.size.Zero() {
		return nil
	}
	var str strings.Builder
	str.WriteString(f2s(v.origin.x))
	str.WriteRune(' ')
	str.WriteString(f2s(v.origin.y))
	str.WriteRune(' ')
	str.WriteString(f2s(v.size.width))
	str.WriteRune(' ')
	str.WriteString(f2s(v.size.height))
	a := xml.NewAttribute(xml.NewName("viewBox"), str.String())
	return []xml.Attribute{a}
}

type Point struct {
	x float64
	y float64
}

func NewPoint(x, y float64) Point {
	return Point{
		x: x,
		y: y,
	}
}

func (p *Point) Move(x, y float64) {
	p.x = x
	p.y = y
}

func (p *Point) Validate() error {
	if err := isFinite(p.x); err != nil {
		return err
	}
	if err := isFinite(p.y); err != nil {
		return err
	}
	return nil
}

func (p *Point) Zero() bool {
	return p.x == 0 && p.y == 0
}

func (p *Point) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("x"), f2s(p.x)),
		xml.NewAttribute(xml.NewName("y"), f2s(p.y)),
	}
}

func (p *Point) centerAttributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("cx"), f2s(p.x)),
		xml.NewAttribute(xml.NewName("cy"), f2s(p.y)),
	}
}

type Size struct {
	width  float64
	height float64
}

func NewSize(width, height float64) Size {
	return Size{
		width:  width,
		height: height,
	}
}

func (s *Size) Resize(width, height float64) {
	s.width = width
	s.height = height
}

func (s *Size) Zero() bool {
	return s.width == 0 && s.height == 0
}

func (s *Size) Validate() error {
	if err := isFinite(s.width); err != nil {
		return err
	}
	if err := isFinite(s.height); err != nil {
		return err
	}
	if s.width < 0 || s.height < 0 {
		return negative("size")
	}
	return nil
}

func (s *Size) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("width"), f2s(s.width)),
		xml.NewAttribute(xml.NewName("height"), f2s(s.height)),
	}
}
