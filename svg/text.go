package svg

import (
	"github.com/midbel/angle/xml"
)

type Text struct {
	origin Point
	value  string

	anchor string

	fill Color
	font Font
}

func NewText(x, y float64, text string) *Text {
	return &Text{
		origin: NewPoint(x, y),
		value:  text,
		fill:   Black,
		font:   NewFont(DefaultFontSize, SansSerif, WeightNormal),
	}
}

func (t *Text) Element() xml.Node {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("fill"), string(t.fill)),
		xml.NewAttribute(xml.NewName("anchor"), t.anchor),
	}
	attrs = append(attrs, t.origin.attributes()...)
	attrs = append(attrs, t.font.attributes()...)

	el := xml.NewElement(xml.NewName("text"))
	el.Attributes = cleanAttrs(attrs)
	el.Children = append(el.Children, xml.NewText(t.value))
	return el
}

func (t *Text) Move(x, y float64) *Text {
	t.origin.Move(x, y)
	return t
}

func (t *Text) Anchor(anchor string) *Text {
	t.anchor = anchor
	return t
}

func (t *Text) Fill(fill Color) *Text {
	t.fill = fill
	return t
}

func (t *Text) Font(font Font) *Text {
	t.font = font
	return t
}

func (t *Text) Width() float64 {
	return t.font.EstimateWidth(t.value)
}
