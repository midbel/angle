package svg

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/midbel/angle/xml"
)

const (
	None  = "none"
	Black = "black"
)

type Color string

func RGB(red, green, blue float64) (Color, error) {
	if red < 0 || red > 255 {
		return "", ErrRange
	}
	if green < 0 || green > 255 {
		return "", ErrRange
	}
	if blue < 0 || blue < 255 {
		return "", ErrRange
	}
	str := fmt.Sprintf("rgb(%s, %s, %s)", f2s(red), f2s(green), f2s(blue))
	return Color(str), nil
}

type Stroke struct {
	width float64
	color string
}

func NewStroke(color string, width float64) Stroke {
	return Stroke{
		width: width,
		color: color,
	}
}

func (s Stroke) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("stroke-width"), f2s(s.width)),
		xml.NewAttribute(xml.NewName("stroke"), s.color),
	}
}

const (
	SansSerif = "sans-serif"
	Helvetica = "helvetica"
)

const (
	WeightNormal  = "normal"
	WeightBold    = "bold"
	WeightBolder  = "bolder"
	WeightLighter = "lighter"
)

type Font struct {
	size   float64
	weight string
	family string
}

func NewFont(size float64, family, weight string) Font {
	return Font{
		size:   size,
		weight: weight,
		family: family,
	}
}

func (f Font) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("font-size"), f2s(f.size)),
		xml.NewAttribute(xml.NewName("font-weight"), f.weight),
		xml.NewAttribute(xml.NewName("font-family"), f.family),
	}
}

const DefaultFontSize = 14

type textFactors struct {
	Default float64
	Space   float64
	Narrow  float64
	Wide    float64
}

func (t textFactors) Coeff(r rune) float64 {
	switch {
	case unicode.IsSpace(r):
		return t.Space
	case strings.ContainsRune("ilIjtfr", r):
		return t.Narrow
	case strings.ContainsRune("MW", r):
		return t.Wide
	default:
		return t.Default
	}
}

var factors = map[string]textFactors{
	SansSerif: {
		Default: 0.53,
		Space:   0.28,
		Narrow:  0.32,
		Wide:    0.88,
	},
	Helvetica: {
		Default: 0.52,
		Space:   0.28,
		Narrow:  0.30,
		Wide:    0.87,
	},
}

func EstimateTextWidthForFont(str, font string, size float64) float64 {
	var (
		coeff = factors[font]
		width float64
	)
	for _, s := range str {
		width += coeff.Coeff(s) * size
	}
	return width
}

func EstimateTextWidth(str string, size float64) float64 {
	return EstimateTextWidthForFont(str, SansSerif, size)
}
