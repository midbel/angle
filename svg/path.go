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
		start:  NewPoint(x1, y1),
		end:    NewPoint(x2, y2),
		stroke: NewStroke(Black, 1),
	}
}

func (i *Line) Element() (xml.Node, error) {
	var attrs []xml.Attribute
	attrs = append(attrs, i.startAttributes()...)
	attrs = append(attrs, i.endAttributes()...)
	attrs = append(attrs, i.stroke.attributes()...)
	attrs = cleanAttrs(attrs)

	el := xml.NewElement(xml.NewName("line"))
	el.Attributes = attrs
	return el, nil
}

func (i *Line) Validate() error {
	return nil
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
	moveTo       pathCmd = 'M'
	lineTo       pathCmd = 'L'
	horizontalTo pathCmd = 'H'
	verticalTo   pathCmd = 'V'
	curveTo      pathCmd = 'C'
	cubicTo      pathCmd = 'S'
	quadraticTo  pathCmd = 'Q'
	arcTo        pathCmd = 'A'
	closePath    pathCmd = 'Z'
)

type step struct {
	Point
	ctrl1 Point
	ctrl2 Point
	cmd pathCmd
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

func (p *Path) Element() (xml.Node, error) {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("d"), p.pathString()),
		xml.NewAttribute(xml.NewName("fill"), p.fill),
	}
	attrs = append(attrs, p.stroke.attributes()...)

	el := xml.NewElement(xml.NewName("path"))
	el.Attributes = cleanAttrs(attrs)
	return el, nil
}

func (p *Path) Validate() error {
	return nil
}

func (p *Path) pathString() string {
	var str strings.Builder

	writePoint := func(pt Point) {
		str.WriteString(f2s(pt.x))
		str.WriteRune(' ')
		str.WriteString(f2s(pt.y))
	}

	for i, s := range p.steps {
		if i > 0 {
			str.WriteRune(' ')
		}
		str.WriteRune(rune(s.cmd))
		if s.cmd == closePath {
			break
		} else if s.cmd == curveTo {
			str.WriteRune(' ')
			writePoint(s.ctrl1)
			str.WriteRune(' ')
			writePoint(s.ctrl2)
			str.WriteRune(' ')
			writePoint(s.Point)
		} else if s.cmd == horizontalTo {
			str.WriteRune(' ')
			str.WriteString(f2s(s.x))
		} else if s.cmd == verticalTo {
			str.WriteRune(' ')
			str.WriteString(f2s(s.y))
		} else {
			str.WriteRune(' ')
			writePoint(s.Point)
		}
	}
	return str.String()
}

func (p *Path) Close() *Path {
	pt := step{
		cmd: closePath,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) Horizontal(x float64) *Path {
	pt := step{
		Point: NewPoint(x, 0),
		cmd:   horizontalTo,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) Vertical(y float64) *Path {
	pt := step{
		Point: NewPoint(0, y),
		cmd:   verticalTo,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) CurveTo(x, y, cx1, cy1, cx2, cy2 float64) *Path {
	pt := step{
		Point: NewPoint(x, y),
		ctrl1: NewPoint(cx1, cy1),
		ctrl2: NewPoint(cx2, cy2),
		cmd:   curveTo,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) MoveTo(x, y float64) *Path {
	pt := step{
		Point: NewPoint(x, y),
		cmd:   moveTo,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) LineTo(x, y float64) *Path {
	pt := step{
		Point: NewPoint(x, y),
		cmd:   lineTo,
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
