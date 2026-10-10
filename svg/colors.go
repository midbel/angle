package svg

import "fmt"

type Color string

const (
	None         Color = "none"
	Transparent  Color = "transparent"
	CurrentColor Color = "currentColor"
)

const (
	AliceBlue            Color = "aliceblue"
	AntiqueWhite         Color = "antiquewhite"
	Aqua                 Color = "aqua"
	Aquamarine           Color = "aquamarine"
	Azure                Color = "azure"
	Beige                Color = "beige"
	Bisque               Color = "bisque"
	Black                Color = "black"
	BlanchedAlmond       Color = "blanchedalmond"
	Blue                 Color = "blue"
	BlueViolet           Color = "blueviolet"
	Brown                Color = "brown"
	Burlywood            Color = "burlywood"
	CadetBlue            Color = "cadetblue"
	Chartreuse           Color = "chartreuse"
	Chocolate            Color = "chocolate"
	Coral                Color = "coral"
	CornflowerBlue       Color = "cornflowerblue"
	Cornsilk             Color = "cornsilk"
	Crimson              Color = "crimson"
	Cyan                 Color = "cyan"
	DarkBlue             Color = "darkblue"
	DarkCyan             Color = "darkcyan"
	DarkGoldenrod        Color = "darkgoldenrod"
	DarkGray             Color = "darkgray"
	DarkGreen            Color = "darkgreen"
	DarkGrey             Color = "darkgrey"
	DarkKhaki            Color = "darkkhaki"
	DarkMagenta          Color = "darkmagenta"
	DarkOliveGreen       Color = "darkolivegreen"
	DarkOrange           Color = "darkorange"
	DarkOrchid           Color = "darkorchid"
	DarkRed              Color = "darkred"
	DarkSalmon           Color = "darksalmon"
	DarkSeaGreen         Color = "darkseagreen"
	DarkSlateBlue        Color = "darkslateblue"
	DarkSlateGray        Color = "darkslategray"
	DarkSlateGrey        Color = "darkslategrey"
	DarkTurquoise        Color = "darkturquoise"
	DarkViolet           Color = "darkviolet"
	DeepPink             Color = "deeppink"
	DeepSkyBlue          Color = "deepskyblue"
	DimGray              Color = "dimgray"
	DimGrey              Color = "dimgrey"
	DodgerBlue           Color = "dodgerblue"
	Firebrick            Color = "firebrick"
	FloralWhite          Color = "floralwhite"
	ForestGreen          Color = "forestgreen"
	Fuchsia              Color = "fuchsia"
	Gainsboro            Color = "gainsboro"
	GhostWhite           Color = "ghostwhite"
	Gold                 Color = "gold"
	Goldenrod            Color = "goldenrod"
	Gray                 Color = "gray"
	Green                Color = "green"
	GreenYellow          Color = "greenyellow"
	Grey                 Color = "grey"
	Honeydew             Color = "honeydew"
	HotPink              Color = "hotpink"
	IndianRed            Color = "indianred"
	Indigo               Color = "indigo"
	Ivory                Color = "ivory"
	Khaki                Color = "khaki"
	Lavender             Color = "lavender"
	LavenderBlush        Color = "lavenderblush"
	LawnGreen            Color = "lawngreen"
	LemonChiffon         Color = "lemonchiffon"
	LightBlue            Color = "lightblue"
	LightCoral           Color = "lightcoral"
	LightCyan            Color = "lightcyan"
	LightGoldenrodYellow Color = "lightgoldenrodyellow"
	LightGray            Color = "lightgray"
	LightGreen           Color = "lightgreen"
	LightGrey            Color = "lightgrey"
	LightPink            Color = "lightpink"
	LightSalmon          Color = "lightsalmon"
	LightSeaGreen        Color = "lightseagreen"
	LightSkyBlue         Color = "lightskyblue"
	LightSlateGray       Color = "lightslategray"
	LightSlateGrey       Color = "lightslategrey"
	LightSteelBlue       Color = "lightsteelblue"
	LightYellow          Color = "lightyellow"
	Lime                 Color = "lime"
	LimeGreen            Color = "limegreen"
	Linen                Color = "linen"
	Magenta              Color = "magenta"
	Maroon               Color = "maroon"
	MediumAquamarine     Color = "mediumaquamarine"
	MediumBlue           Color = "mediumblue"
	MediumOrchid         Color = "mediumorchid"
	MediumPurple         Color = "mediumpurple"
	MediumSeaGreen       Color = "mediumseagreen"
	MediumSlateBlue      Color = "mediumslateblue"
	MediumSpringGreen    Color = "mediumspringgreen"
	MediumTurquoise      Color = "mediumturquoise"
	MediumVioletRed      Color = "mediumvioletred"
	MidnightBlue         Color = "midnightblue"
	MintCream            Color = "mintcream"
	MistyRose            Color = "mistyrose"
	Moccasin             Color = "moccasin"
	NavajoWhite          Color = "navajowhite"
	Navy                 Color = "navy"
	OldLace              Color = "oldlace"
	Olive                Color = "olive"
	OliveDrab            Color = "olivedrab"
	Orange               Color = "orange"
	OrangeRed            Color = "orangered"
	Orchid               Color = "orchid"
	PaleGoldenrod        Color = "palegoldenrod"
	PaleGreen            Color = "palegreen"
	PaleTurquoise        Color = "paleturquoise"
	PaleVioletRed        Color = "palevioletred"
	PapayaWhip           Color = "papayawhip"
	PeachPuff            Color = "peachpuff"
	Peru                 Color = "peru"
	Pink                 Color = "pink"
	Plum                 Color = "plum"
	PowderBlue           Color = "powderblue"
	Purple               Color = "purple"
	RebeccaPurple        Color = "rebeccapurple"
	Red                  Color = "red"
	RosyBrown            Color = "rosybrown"
	RoyalBlue            Color = "royalblue"
	SaddleBrown          Color = "saddlebrown"
	Salmon               Color = "salmon"
	SandyBrown           Color = "sandybrown"
	SeaGreen             Color = "seagreen"
	Seashell             Color = "seashell"
	Sienna               Color = "sienna"
	Silver               Color = "silver"
	SkyBlue              Color = "skyblue"
	SlateBlue            Color = "slateblue"
	SlateGray            Color = "slategray"
	SlateGrey            Color = "slategrey"
	Snow                 Color = "snow"
	SpringGreen          Color = "springgreen"
	SteelBlue            Color = "steelblue"
	Tan                  Color = "tan"
	Teal                 Color = "teal"
	Thistle              Color = "thistle"
	Tomato               Color = "tomato"
	Turquoise            Color = "turquoise"
	Violet               Color = "violet"
	Wheat                Color = "wheat"
	White                Color = "white"
	WhiteSmoke           Color = "whitesmoke"
	Yellow               Color = "yellow"
	YellowGreen          Color = "yellowgreen"
)

func RGB(red, green, blue int) (Color, error) {
	if red < 0 || red > 255 {
		return None, outOfRange("red")
	}
	if green < 0 || green > 255 {
		return None, outOfRange("green")
	}
	if blue < 0 || blue > 255 {
		return None, outOfRange("blue")
	}
	str := fmt.Sprintf("rgb(%d, %d, %d)", red, green, blue)
	return Color(str), nil
}

func RGBA(red, green, blue int, alpha float64) (Color, error) {
	if red < 0 || red > 255 {
		return None, outOfRange("red")
	}
	if green < 0 || green > 255 {
		return None, outOfRange("green")
	}
	if blue < 0 || blue > 255 {
		return None, outOfRange("blue")
	}
	if err := isFinite(alpha); err != nil {
		return None, err
	}
	if alpha < 0 || alpha > 1 {
		return None, outOfRange("alpha")
	}
	str := fmt.Sprintf("rgba(%d, %d, %d, %f)", red, green, blue, alpha)
	return Color(str), nil
}

func Hex(s string) Color {
	return Color(s)
}
