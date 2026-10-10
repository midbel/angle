package svg

func EstimateTextWidthForFont(text string, font string, size float64) float64 {
	var (
		metrics = resolveMetrics(font)
		width   float64
	)
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
	return spacing * defaultSpacingWidth * size
}

const (
	defaultGlyphWidth   = 0.6
	defaultSpacingWidth = 0.5
)

var defaultGlyphMetrics = map[rune]float64{
	// Lowercase
	'a': 0.54, 'b': 0.56, 'c': 0.48, 'd': 0.56,
	'e': 0.54, 'f': 0.30, 'g': 0.56, 'h': 0.56,
	'i': 0.24, 'j': 0.24, 'k': 0.52, 'l': 0.24,
	'm': 0.86, 'n': 0.56, 'o': 0.56, 'p': 0.56,
	'q': 0.56, 'r': 0.36, 's': 0.48, 't': 0.32,
	'u': 0.56, 'v': 0.50, 'w': 0.74, 'x': 0.50,
	'y': 0.50, 'z': 0.48,

	// latin
	'é': 0.54, 'è': 0.54, 'ê': 0.54, 'ë': 0.54,
	'à': 0.54,
	'ù': 0.54,
	'ô': 0.56,
	'î': 0.24, 'ï': 0.24,
	'ç': 0.48,

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

type FontCategory uint8

const (
	FontSansSerif FontCategory = iota
	FontSerif
	FontMonospace
	FontCursive
	FontFantasy
)

var fontCategories = map[string]FontCategory{
	Arial:         FontSansSerif,
	Helvetica:     FontSansSerif,
	Verdana:       FontSansSerif,
	Tahoma:        FontSansSerif,
	TrebuchetMS:   FontSansSerif,
	TimesNewRoman: FontSerif,
	Georgia:       FontSerif,
	CourierNew:    FontMonospace,

	SansSerif: FontSansSerif,
	Serif:     FontSerif,
	Monospace: FontMonospace,
	Cursive:   FontCursive,
	Fantasy:   FontFantasy,
	SystemUI:  FontSansSerif,
}

func resolveMetrics(family string) fontMetrics {
	if m, ok := defaultFontMetrics[family]; ok {
		return m
	}

	category, ok := fontCategories[family]
	if !ok {
		category = FontSansSerif
	}

	switch category {
	case FontSerif:
		return defaultFontMetrics[Serif]
	case FontMonospace:
		return defaultFontMetrics[Monospace]
	default:
		return defaultFontMetrics[SansSerif]
	}
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
