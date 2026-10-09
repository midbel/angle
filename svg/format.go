package svg

import (
	"slices"
	"strings"
	"strconv"

	"github.com/midbel/angle/xml"
)

func f2s(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func af2s(values []float64) string {
	list := make([]string, len(values))
	for i := range values {
		list[i] = f2s(values[i])
	}
	return strings.Join(list, " ")
}

func cleanAttrs(attrs []xml.Attribute) []xml.Attribute {
	return slices.DeleteFunc(attrs, func(a xml.Attribute) bool {
		return a.Value == ""
	})
}
