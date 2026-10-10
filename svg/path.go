package svg

import (
	"fmt"
	"slices"
	"strings"

	"github.com/midbel/angle/xml"
)

type Line struct {
	*Attributes

	start Point
	end   Point

	stroke Stroke
}

func NewLine(x1, y1, x2, y2 float64) *Line {
	return &Line{
		Attributes: NewAttributes(),
		start:      NewPoint(x1, y1),
		end:        NewPoint(x2, y2),
		stroke:     NewStroke(Black, 1),
	}
}

func (i *Line) Element() (xml.Node, error) {
	if err := i.Validate(); err != nil {
		return nil, err
	}
	var attrs []xml.Attribute
	attrs = append(attrs, i.startAttributes()...)
	attrs = append(attrs, i.endAttributes()...)
	attrs = append(attrs, i.stroke.attributes()...)
	attrs = append(attrs, i.Attributes.attributes()...)
	attrs = cleanAttrs(attrs)

	el := xml.NewElement(xml.NewName("line"))
	el.Attributes = attrs
	return el, nil
}

func (i *Line) Validate() error {
	if err := i.start.Validate(); err != nil {
		return err
	}
	if err := i.end.Validate(); err != nil {
		return err
	}
	if err := i.stroke.Validate(); err != nil {
		return err
	}
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

type Polyline struct {
	*Attributes

	points []Point
	fill   Color
	stroke Stroke
}

func NewPolyline(points ...Point) Polyline {
	return Polyline{
		Attributes: NewAttributes(),
		points:     slices.Clone(points),
	}
}

func (p Polyline) Fill(fill Color) Polyline {
	p.fill = fill
	return p
}

func (p Polyline) Stroke(stroke Stroke) Polyline {
	p.stroke = stroke
	return p
}

func (p Polyline) Validate() error {
	for i := range p.points {
		if err := p.points[i].Validate(); err != nil {
			return err
		}
	}
	if err := p.stroke.Validate(); err != nil {
		return err
	}
	return nil
}

func (p Polyline) Element() (xml.Node, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("points"), p.getPoints()),
		xml.NewAttribute(xml.NewName("fill"), string(p.fill)),
	}
	attrs = append(attrs, p.stroke.attributes()...)
	el := xml.NewElement(xml.NewName("polyline"))
	el.Attributes = cleanAttrs(attrs)
	return el, nil
}

func (p Polyline) getPoints() string {
	var str strings.Builder
	for i := range p.points {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(f2s(p.points[i].x))
		str.WriteRune(' ')
		str.WriteString(f2s(p.points[i].y))
	}
	return str.String()
}

type Path struct {
	*Attributes

	steps []command

	fill   Color
	stroke Stroke
}

func NewPath() *Path {
	return &Path{
		Attributes: NewAttributes(),
		stroke:     NewStroke(Black, 1),
		fill:       None,
	}
}

func (p *Path) Element() (xml.Node, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("d"), p.buildPath()),
		xml.NewAttribute(xml.NewName("fill"), string(p.fill)),
	}
	attrs = append(attrs, p.stroke.attributes()...)
	attrs = append(attrs, p.Attributes.attributes()...)

	el := xml.NewElement(xml.NewName("path"))
	el.Attributes = cleanAttrs(attrs)
	return el, nil
}

func (p *Path) Validate() error {
	if err := p.stroke.Validate(); err != nil {
		return err
	}
	if len(p.steps) == 0 {
		return fmt.Errorf("empty path - no commands")
	}
	if p.steps[0].Command() != cmdMoveTo {
		return fmt.Errorf("invalid path - no move command")
	}
	return nil
}

func (p *Path) buildPath() string {
	var str strings.Builder

	for i, s := range p.steps {
		if i > 0 {
			str.WriteRune(' ')
		}
		str.WriteRune(rune(s.Command()))
		str.WriteRune(' ')
		str.WriteString(s.String())
	}
	return str.String()
}

func (p *Path) Close() *Path {
	var pt closeTo
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) Horizontal(x float64) *Path {
	pt := horizontalTo{
		Point: NewPoint(x, 0),
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) Vertical(y float64) *Path {
	pt := verticalTo{
		Point: NewPoint(0, y),
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) CurveTo(x, y, cx1, cy1, cx2, cy2 float64) *Path {
	pt := curveTo{
		Point: NewPoint(x, y),
		ctrl1: NewPoint(cx1, cy1),
		ctrl2: NewPoint(cx2, cy2),
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) CubicTo(x, y, cx, cy float64) *Path {
	pt := cubicTo{
		Point: NewPoint(x, y),
		ctrl:  NewPoint(cx, cy),
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) QuadraticTo(x, y, cx, cy float64) *Path {
	pt := quadraticTo{
		Point: NewPoint(x, y),
		ctrl:  NewPoint(cx, cy),
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) ArcTo(x, y, rx, ry, rotation float64, large, sweep bool) *Path {
	pt := arcTo{
		Point:     NewPoint(x, y),
		rotation:  rotation,
		radius:    NewPoint(rx, ry),
		largeArc:  large,
		sweepFlag: sweep,
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) MoveTo(x, y float64) *Path {
	pt := moveTo{
		Point: NewPoint(x, y),
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) LineTo(x, y float64) *Path {
	pt := lineTo{
		Point: NewPoint(x, y),
	}
	p.steps = append(p.steps, pt)
	return p
}

func (p *Path) Fill(fill Color) *Path {
	p.fill = fill
	return p
}

func (p *Path) Stroke(stroke Stroke) *Path {
	p.stroke = stroke
	return p
}

type pathCmd rune

const (
	cmdMoveTo       pathCmd = 'M'
	cmdLineTo       pathCmd = 'L'
	cmdHorizontalTo pathCmd = 'H'
	cmdVerticalTo   pathCmd = 'V'
	cmdCurveTo      pathCmd = 'C'
	cmdCubicTo      pathCmd = 'S'
	cmdQuadraticTo  pathCmd = 'Q'
	cmdArcTo        pathCmd = 'A'
	cmdCloseTo      pathCmd = 'Z'
)

type command interface {
	String() string
	Command() pathCmd
}

type moveTo struct {
	Point
}

func (c moveTo) String() string {
	return writePoint(c.Point)
}

func (moveTo) Command() pathCmd {
	return cmdMoveTo
}

type lineTo struct {
	Point
}

func (c lineTo) String() string {
	return writePoint(c.Point)
}

func (lineTo) Command() pathCmd {
	return cmdLineTo
}

type horizontalTo struct {
	Point
}

func (c horizontalTo) String() string {
	return writePoint(c.Point)
}

func (horizontalTo) Command() pathCmd {
	return cmdHorizontalTo
}

type verticalTo struct {
	Point
}

func (c verticalTo) String() string {
	return writePoint(c.Point)
}

func (verticalTo) Command() pathCmd {
	return cmdVerticalTo
}

type curveTo struct {
	Point
	ctrl1 Point
	ctrl2 Point
}

func (c curveTo) String() string {
	var str strings.Builder
	str.WriteString(writePoint(c.ctrl1))
	str.WriteRune(' ')
	str.WriteString(writePoint(c.ctrl2))
	str.WriteRune(' ')
	str.WriteString(writePoint(c.Point))
	return str.String()
}

func (curveTo) Command() pathCmd {
	return cmdCurveTo
}

type cubicTo struct {
	Point
	ctrl Point
}

func (c cubicTo) String() string {
	var str strings.Builder
	str.WriteString(writePoint(c.ctrl))
	str.WriteRune(' ')
	str.WriteString(writePoint(c.Point))
	return str.String()
}

func (cubicTo) Command() pathCmd {
	return cmdCubicTo
}

type quadraticTo struct {
	Point
	ctrl Point
}

func (c quadraticTo) String() string {
	var str strings.Builder
	str.WriteString(writePoint(c.ctrl))
	str.WriteRune(' ')
	str.WriteString(writePoint(c.Point))
	return str.String()
}

func (quadraticTo) Command() pathCmd {
	return cmdQuadraticTo
}

type arcTo struct {
	Point
	radius    Point
	rotation  float64
	largeArc  bool
	sweepFlag bool
}

func (c arcTo) String() string {
	var str strings.Builder
	str.WriteString(writePoint(c.radius))
	str.WriteRune(' ')
	str.WriteString(f2s(c.rotation))
	str.WriteRune(' ')
	str.WriteString(b2s(c.largeArc))
	str.WriteRune(' ')
	str.WriteString(b2s(c.largeArc))
	str.WriteRune(' ')
	str.WriteString(writePoint(c.Point))
	return str.String()
}

func (arcTo) Command() pathCmd {
	return cmdArcTo
}

type closeTo struct{}

func (c closeTo) String() string {
	return ""
}

func (closeTo) Command() pathCmd {
	return cmdCloseTo
}

func writePoint(pt Point) string {
	var str strings.Builder
	str.WriteString(f2s(pt.x))
	str.WriteRune(' ')
	str.WriteString(f2s(pt.y))
	return str.String()
}
