package svg

import (
	"fmt"

	"github.com/midbel/angle/xml"
)

const (
	Black Color = "black"
	None  Color = "none"
)

type Color string

func RGB(red, green, blue int) (Color, error) {
	if red < 0 || red > 255 {
		return "", outOfRange("red")
	}
	if green < 0 || green > 255 {
		return "", outOfRange("green")
	}
	if blue < 0 || blue > 255 {
		return "", outOfRange("blue")
	}
	str := fmt.Sprintf("rgb(%d, %d, %d)", red, green, blue)
	return Color(str), nil
}

type Stroke struct {
	width    float64
	opacity  float64
	array    []float64
	color    Color
	lineCap  string
	lineJoin string
}

func NewStroke(color Color, width float64) Stroke {
	return Stroke{
		width: width,
		color: color,
		opacity: 1,
	}
}

func (s Stroke) Validate() error {
	if s.width < 0 {
		return negative("stroke-width")
	}
	return nil
}

func (s Stroke) Opacity(val float64) Stroke {
	s.opacity = val
	return s
}

func (s Stroke) Array(n ...float64) Stroke {
	s.array = n[:]
	return s
}

func (s Stroke) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("stroke-width"), f2s(s.width)),
		xml.NewAttribute(xml.NewName("stroke-opacity"), f2s(s.opacity)),
		xml.NewAttribute(xml.NewName("stroke-dasharray"), af2s(s.array)),
		xml.NewAttribute(xml.NewName("stroke"), string(s.color)),
		xml.NewAttribute(xml.NewName("stroke-linecap"), string(s.lineCap)),
		xml.NewAttribute(xml.NewName("stroke-linejoin"), string(s.lineJoin)),
	}
}

const (
	SansSerif = "sans-serif"
	Serif = "serif"
	Monospace = "monospace"
	Helvetica = "helvetica"
)

const (
	WeightNormal  = "normal"
	WeightBold    = "bold"
	WeightBolder  = "bolder"
	WeightLighter = "lighter"
)

const defaultSize = 12

type Font struct {
	size      float64
	weight    string
	family    string
	style     string
	underline bool
	striked   bool
}

func NewFont(size float64, family, weight string) Font {
	if size == 0 {
		size = defaultSize
	}
	return Font{
		size:   size,
		weight: weight,
		family: family,
		style:  "normal",
	}
}

func (f Font) Validate() error {
	if f.size < 0 {
		return negative("font-size")
	}
	return nil
}

func (f Font) Italic() Font {
	f.style = "italic"
	return f
}

func (f Font) Normal() Font {
	f.style = "normal"
	return f
}

func (f Font) Underline() Font {
	f.underline = true
	return f
}

func (f Font) StrikeThrought() Font {
	f.striked = true
	return f
}

func (f Font) attributes() []xml.Attribute {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("font-size"), f2s(f.size)),
		xml.NewAttribute(xml.NewName("font-weight"), f.weight),
		xml.NewAttribute(xml.NewName("font-family"), f.family),
		xml.NewAttribute(xml.NewName("font-style"), f.style),
	}
	if f.underline {
		a := xml.NewAttribute(xml.NewName("text-decoration"), "underline")
		attrs = append(attrs, a)
	}
	if f.striked {
		a := xml.NewAttribute(xml.NewName("text-decoration"), "line-throught")
		attrs = append(attrs, a)
	}
	return attrs
}

func (f Font) EstimateWidth(text string) float64 {
	return EstimateTextWidthForFont(text, f.family, f.size)
}

const DefaultFontSize = 14

func EstimateTextWidthForFont(text string, font string, size float64) float64 {
	metrics, ok := defaultFontMetrics[font]
	if !ok {
		metrics = defaultFontMetrics[SansSerif]
	}

	var width float64
	for _, r := range text {
		w, ok := defaultGlyphMetrics[r]
		if !ok {
			w = defaultGlyphWidth
		}

		if o, ok := metrics.Overrides[r]; ok {
			w = o
		}

		if metrics.FixedWidth > 0 {
			w = metrics.FixedWidth
		}

		width += w*metrics.Scale + metrics.Adjust
	}
	return width * size
}

func EstimateTextWidth(str string, size float64) float64 {
	return EstimateTextWidthForFont(str, SansSerif, size)
}

func EstimateSpacingWidth(spacing, size float64) float64 {
    return spacing * 0.5 * size
}

var defaultGlyphWidth = 0.6

var defaultGlyphMetrics = map[rune]float64{
	// Lowercase
	'a': 0.54, 'b': 0.56, 'c': 0.48, 'd': 0.56,
	'e': 0.54, 'f': 0.30, 'g': 0.56, 'h': 0.56,
	'i': 0.24, 'j': 0.24, 'k': 0.52, 'l': 0.24,
	'm': 0.86, 'n': 0.56, 'o': 0.56, 'p': 0.56,
	'q': 0.56, 'r': 0.36, 's': 0.48, 't': 0.32,
	'u': 0.56, 'v': 0.50, 'w': 0.74, 'x': 0.50,
	'y': 0.50, 'z': 0.48,

	// Uppercase
	'A': 0.68, 'B': 0.63, 'C': 0.69, 'D': 0.72,
	'E': 0.58, 'F': 0.54, 'G': 0.74, 'H': 0.72,
	'I': 0.28, 'J': 0.50, 'K': 0.65, 'L': 0.54,
	'M': 0.86, 'N': 0.72, 'O': 0.76, 'P': 0.60,
	'Q': 0.76, 'R': 0.66, 'S': 0.60, 'T': 0.58,
	'U': 0.70, 'V': 0.68, 'W': 0.96, 'X': 0.62,
	'Y': 0.62, 'Z': 0.58,

	// Digits
	'0': 0.56, '1': 0.56, '2': 0.56, '3': 0.56,
	'4': 0.56, '5': 0.56, '6': 0.56, '7': 0.56,
	'8': 0.56, '9': 0.56,

	// Whitespace and punctuation
	' ':  0.28,
	'\t': 1.12, // Convention provisoire : 4 spaces
	'.':  0.26, ',': 0.26, ':': 0.26, ';': 0.26,
	'!': 0.28, '?': 0.52,
	'\'': 0.20, '"': 0.36,
	'-': 0.34, '_': 0.56,
	'(': 0.32, ')': 0.32,
	'[': 0.32, ']': 0.32,
	'{': 0.36, '}': 0.36,
	'/': 0.32, '\\': 0.32,
	'+': 0.60, '=': 0.60,
	'*': 0.42, '&': 0.66,
	'%': 0.86, '#': 0.60,
	'@': 0.90,
}

type fontMetrics struct {
	Scale      float64
	Adjust     float64
	FixedWidth float64
	Overrides  map[rune]float64
}

var defaultFontMetrics = map[string]fontMetrics{
	SansSerif: {
		Scale:  1.00,
		Adjust: 0.00,
	},
	Serif: {
		Scale:  1.02,
		Adjust: 0.00,
		Overrides: map[rune]float64{
			'i': 0.28,
			'l': 0.28,
			'm': 0.83,
			'W': 0.94,
		},
	},
	Monospace: {
		Scale:      1.00,
		Adjust:     0.00,
		FixedWidth: 0.60,
	},
}
