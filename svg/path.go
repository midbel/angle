package svg

import (
	"strings"

	"github.com/midbel/angle/xml"
)

type moveType rune

const (
	moveTo = 'M'
	lineTo = 'L'
)

type point struct {
	X    float64
	Y    float64
	move moveType
}

type Path struct {
	points []point

	fill        string
	stroke      string
	strokeWidth float64
}

func NewPath() *Path {
	return &Path{
		stroke:      Black,
		strokeWidth: 1,
		fill:        None,
	}
}

func (p *Path) Element() xml.Node {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("d"), p.pathString()),
		xml.NewAttribute(xml.NewName("fill"), p.fill),
		xml.NewAttribute(xml.NewName("stroke"), p.stroke),
	}
	if p.strokeWidth > 0 {
		a := xml.NewAttribute(xml.NewName("stroke-width"), f2s(p.strokeWidth))
		attrs = append(attrs, a)
	}
	el := xml.NewElement(xml.NewName("path"))
	el.Attributes = cleanAttrs(attrs)
	return el
}

func (p *Path) pathString() string {
	var str strings.Builder
	for i, p := range p.points {
		if i > 0 {
			str.WriteRune(' ')
		}
		str.WriteRune(rune(p.move))
		str.WriteRune(' ')
		str.WriteString(f2s(p.X))
		str.WriteRune(' ')
		str.WriteString(f2s(p.Y))
	}
	return str.String()
}

func (p *Path) MoveTo(x, y float64) *Path {
	pt := point{
		X:    x,
		Y:    y,
		move: moveTo,
	}
	p.points = append(p.points, pt)
	return p
}

func (p *Path) LineTo(x, y float64) *Path {
	pt := point{
		X:    x,
		Y:    y,
		move: lineTo,
	}
	p.points = append(p.points, pt)
	return p
}

func (p *Path) Fill(fill string) *Path {
	p.fill = fill
	return p
}

func (p *Path) Stroke(stroke string) *Path {
	p.stroke = stroke
	return p
}

func (p *Path) StrokeWidth(width float64) *Path {
	p.strokeWidth = width
	return p
}
