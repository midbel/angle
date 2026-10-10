package svg

import (
	"slices"
	"strings"

	"github.com/midbel/angle/xml"
)

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
		width:   width,
		color:   color,
		opacity: 1,
	}
}

func (s Stroke) Validate() error {
	if err := isFinite(s.width); err != nil {
		return err
	}
	if err := isFinite(s.opacity); err != nil {
		return err
	}
	if s.opacity < 0 || s.opacity > 1 {
		return outOfRange("stroke-opacity")
	}
	for i := range s.array {
		if err := isFinite(s.array[i]); err != nil {
			return err
		}
		if s.array[i] < 0 {
			return negative("stroke-dasharray")
		}
	}
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
	s.array = slices.Clone(n)
	return s
}

func (s Stroke) LineCap(val string) Stroke {
	s.lineCap = val
	return s
}

func (s Stroke) LineJoin(val string) Stroke {
	s.lineJoin = val
	return s
}

func (s Stroke) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("stroke-width"), f2s(s.width)),
		xml.NewAttribute(xml.NewName("stroke-opacity"), f2s(s.opacity)),
		xml.NewAttribute(xml.NewName("stroke-dasharray"), af2s(s.array)),
		xml.NewAttribute(xml.NewName("stroke"), string(s.color)),
		xml.NewAttribute(xml.NewName("stroke-linecap"), s.lineCap),
		xml.NewAttribute(xml.NewName("stroke-linejoin"), s.lineJoin),
	}
}

const (
	// Generic font families
	Serif     = "serif"
	SansSerif = "sans-serif"
	Monospace = "monospace"
	Cursive   = "cursive"
	Fantasy   = "fantasy"

	// Common font families
	Arial         = "Arial"
	Helvetica     = "Helvetica"
	TimesNewRoman = "Times New Roman"
	Georgia       = "Georgia"
	Verdana       = "Verdana"
	Tahoma        = "Tahoma"
	TrebuchetMS   = "Trebuchet MS"
	CourierNew    = "Courier New"
	SystemUI      = "system-ui"
)

const (
	WeightNormal  = "normal"
	WeightBold    = "bold"
	WeightBolder  = "bolder"
	WeightLighter = "lighter"
)

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
		size = DefaultFontSize
	}
	return Font{
		size:   size,
		weight: weight,
		family: family,
		style:  "normal",
	}
}

func (f Font) Validate() error {
	if err := isFinite(f.size); err != nil {
		return err
	}
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

func (f Font) StrikeThrough() Font {
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
	var decorations []string
	if f.underline {
		decorations = append(decorations, "underline")
	}
	if f.striked {
		decorations = append(decorations, "line-through")
	}
	if len(decorations) > 0 {
		a := xml.NewAttribute(xml.NewName("text-decoration"), strings.Join(decorations, " "))
		attrs = append(attrs, a)
	}
	return attrs
}

func (f Font) EstimateWidth(text string) float64 {
	return EstimateTextWidthForFont(text, f.family, f.size)
}

const DefaultFontSize = 14
