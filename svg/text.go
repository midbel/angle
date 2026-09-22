package svg

import (
	"github.com/midbel/angle/xml"
)

type Text struct {
	x     float64
	y     float64
	value string

	anchor string

	fill       string
	fontSize   float64
	fontWeight float64
	fontFamily string
}

func NewText(x, y float64, text string) *Text {
	return &Text{
		x:     x,
		y:     y,
		value: text,
		fill:  Black,
	}
}

func (t *Text) Element() xml.Node {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("x"), f2s(t.x)),
		xml.NewAttribute(xml.NewName("y"), f2s(t.y)),
		xml.NewAttribute(xml.NewName("fill"), t.fill),
		xml.NewAttribute(xml.NewName("font-family"), t.fontFamily),
	}
	if t.fontSize > 0 {
		a := xml.NewAttribute(xml.NewName("font-size"), f2s(t.fontSize))
		attrs = append(attrs, a)
	}
	if t.fontWeight > 0 {
		a := xml.NewAttribute(xml.NewName("font-weight"), f2s(t.fontWeight))
		attrs = append(attrs, a)
	}
	el := xml.NewElement(xml.NewName("text"))
	el.Attributes = cleanAttrs(attrs)
	el.Children = append(el.Children, xml.NewText(t.value))
	return el
}

func (t *Text) Move(x, y float64) *Text {
	t.x = x
	t.y = y
	return t
}

func (t *Text) Anchor(anchor string) *Text {
	t.anchor = anchor
	return t
}

func (t *Text) Fill(fill string) *Text {
	t.fill = fill
	return t
}

func (t *Text) FontSize(size float64) *Text {
	t.fontSize = size
	return t
}

func (t *Text) FontWeight(weight float64) *Text {
	t.fontWeight = weight
	return t
}

func (t *Text) FontFamily(font string) *Text {
	t.fontFamily = font
	return t
}
