package svg

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/midbel/angle/xml"
)

const None = "none"

const (
	Black Color = "black"
)

type Color string

func RGB(red, green, blue int) (Color, error) {
	if red < 0 || red > 255 {
		return "", outOfRange("red")
	}
	if green < 0 || green > 255 {
		return "", outOfRange("red")
	}
	if blue < 0 || blue > 255 {
		return "", outOfRange("red")
	}
	str := fmt.Sprintf("rgb(%d, %d, %d)", red, green, blue)
	return Color(str), nil
}

type Stroke struct {
	width float64
	color Color
}

func NewStroke(color Color, width float64) Stroke {
	return Stroke{
		width: width,
		color: color,
	}
}

func (s Stroke) Validate() error {
	if s.width < 0 {
		return negative("stroke-width")
	}
	return nil
}

func (s Stroke) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("stroke-width"), f2s(s.width)),
		xml.NewAttribute(xml.NewName("stroke"), string(s.color)),
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

func (f Font) Validate() error {
	if f.size < 0 {
		return negative("font-size")
	}
	return nil
}

func (f Font) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("font-size"), f2s(f.size)),
		xml.NewAttribute(xml.NewName("font-weight"), f.weight),
		xml.NewAttribute(xml.NewName("font-family"), f.family),
	}
}

func (f Font) EstimateWidth(text string) float64 {
	return EstimateTextWidthForFont(text, f.family, f.size)
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
