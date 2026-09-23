package svg

import (
	"strings"

	"github.com/midbel/angle/xml"
)

type Line struct {
	start Point
	end   Point

	stroke Stroke
}

func NewLine(x1, y1, x2, y2 float64) *Line {
	return &Line{
		start: NewPoint(x1, y1),
		end:   NewPoint(x2, y2),
		stroke: NewStroke(Black, 1),
	}
}

func (i *Line) Element() xml.Node {
	var attrs []xml.Attribute
	attrs = append(attrs, i.startAttributes()...)
	attrs = append(attrs, i.endAttributes()...)
	attrs = append(attrs, i.stroke.attributes()...)
	attrs = cleanAttrs(attrs)

	el := xml.NewElement(xml.NewName("line"))
	el.Attributes = attrs
	return el
}

func (i *Line) startAttributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("x1"), f2s(i.start.x)),
		xml.NewAttribute(xml.NewName("y1"), f2s(i.start.y)),
	}
}

func (i *Line) endAttributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("x2"), f2s(i.end.x)),
		xml.NewAttribute(xml.NewName("y2"), f2s(i.end.y)),
	}
}

type pathCmd rune

const (
	moveTo pathCmd = 'M'
	lineTo pathCmd = 'L'
)

type step struct {
	Point
	move pathCmd
}

type Path struct {
	steps []step

	fill   string
	stroke Stroke
}

func NewPath() *Path {
	return &Path{
		stroke: NewStroke(Black, 1),
		fill:   None,
	}
}

func (p *Path) Element() xml.Node {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("d"), p.pathString()),
		xml.NewAttribute(xml.NewName("fill"), p.fill),
	}
	attrs = append(attrs, p.stroke.attributes()...)

	el := xml.NewElement(xml.NewName("path"))
	el.Attributes = cleanAttrs(attrs)
	return el
}

func (p *Path) pathString() string {
	var str strings.Builder
	for i, s := range p.steps {
		if i > 0 {
			str.WriteRune(' ')
		}
		str.WriteRune(rune(s.move))
		str.WriteRune(' ')
		str.WriteString(f2s(s.x))
		str.WriteRune(' ')
		str.WriteString(f2s(s.y))
	}
	return str.String()
}

func (p *Path) MoveTo(x, y float64) *Path {
	pt := step{
		Point: NewPoint(x, y),
		move:  moveTo,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) LineTo(x, y float64) *Path {
	pt := step{
		Point: NewPoint(x, y),
		move:  lineTo,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) Fill(fill string) *Path {
	p.fill = fill
	return p
}

func (p *Path) Stroke(stroke Stroke) *Path {
	p.stroke = stroke
	return p
}
