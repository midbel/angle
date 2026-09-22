package svg

import (
	"slices"
	"strconv"

	"github.com/midbel/angle/xml"
)

func f2s(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func cleanAttrs(attrs []xml.Attribute) []xml.Attribute {
	return slices.DeleteFunc(attrs, func(a xml.Attribute) bool {
		return a.Value == ""
	})
}
