package svg

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/midbel/angle/xml"
)

type Attributes struct {
	id     string
	class  []string
	styles map[string]string
}

func NewAttributes() *Attributes {
	a := &Attributes{
		styles: make(map[string]string),
	}
	return a
}

func (a *Attributes) Id(id string) *Attributes {
	a.id = id
	return a
}

func (a *Attributes) Class(class []string) *Attributes {
	a.class = class
	return a
}

func (a *Attributes) Style(name, value string) *Attributes {
	a.styles[name] = value
	return a
}

func (a *Attributes) attributes() []xml.Attribute {
	attrs := []xml.Attribute{
		xml.NewAttribute(xml.NewName("id"), a.id),
		xml.NewAttribute(xml.NewName("class"), strings.Join(a.class, " ")),
		xml.NewAttribute(xml.NewName("style"), a.getStyles()),
	}
	return attrs
}

func (a *Attributes) getStyles() string {
	var (
		keys = slices.Sorted(maps.Keys(a.styles))
		str  strings.Builder
	)
	for _, k := range keys {
		v := a.styles[k]
		if v == "" {
			continue
		}
		str.WriteString(k)
		str.WriteRune(':')
		str.WriteString(v)
		str.WriteRune(';')
	}
	return str.String()
}

type translate struct {
	left float64
	top  float64
}

func (t translate) String() string {
	return fmt.Sprintf("translate(%f, %f)", t.left, t.top)
}

type rotate struct {
	angle  float64
	origin Point
}

func (r rotate) String() string {
	return fmt.Sprintf("rotate(%f)", r.angle)
}

type scale struct {
	x float64
	y float64
}

func (s scale) String() string {
	return fmt.Sprintf("scale(%f, %f)", s.x, s.y)
}

type skewX struct {
	degree float64
}

func (x skewX) String() string {
	return fmt.Sprintf("skewX(%f)", x.degree)
}

type skewY struct {
	degree float64
}

func (x skewY) String() string {
	return fmt.Sprintf("skewY(%f)", x.degree)
}

var (
	ErrNegative = errors.New("negative value")
	ErrRange    = errors.New("out of range")
	ErrNumber   = errors.New("invalid number")
)

func negative(field string) error {
	return fmt.Errorf("%s: can not have a %w", field, ErrNegative)
}

func outOfRange(field string) error {
	return fmt.Errorf("%s: value is %w", field, ErrRange)
}
