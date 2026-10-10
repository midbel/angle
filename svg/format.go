package svg

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/midbel/angle/xml"
)

func f2s(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func b2s(val bool) string {
	if val {
		return "1"
	}
	return "0"
}

func af2s(values []float64) string {
	list := make([]string, len(values))
	for i := range values {
		list[i] = f2s(values[i])
	}
	return strings.Join(list, " ")
}

func isFinite(f float64) error {
	if ok := math.IsInf(f, 0); ok {
		return fmt.Errorf("%w: infinite not allowed", ErrNumber)
	}
	if ok := math.IsNaN(f); ok {
		return fmt.Errorf("%w: not a number", ErrNumber)
	}
	return nil
}

func cleanAttrs(attrs []xml.Attribute) []xml.Attribute {
	return slices.DeleteFunc(attrs, func(a xml.Attribute) bool {
		return a.Value == ""
	})
}
