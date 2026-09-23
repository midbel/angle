package svg

import (
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
	if r.x < 0 || r.y < 0 {
		return ErrNegative
	}
	return nil
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
	if p.x < 0 || p.y < 0 {
		return ErrNegative
	}
	return nil
}

func (p *Point) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("x"), f2s(p.x)),
		xml.NewAttribute(xml.NewName("y"), f2s(p.y)),
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

func (s *Size) Validate() error {
	if s.width < 0 || s.height < 0 {
		return ErrNegative
	}
	return nil
}

func (s *Size) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("width"), f2s(s.width)),
		xml.NewAttribute(xml.NewName("height"), f2s(s.height)),
	}
}
