package svg

import (
	"fmt"
	"errors"
	"strings"

	"github.com/midbel/angle/xml"
)

type Attributes struct {
	id        string
	class     []string
	transform string
	styles    map[string]string
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

func (a *Attributes) Translate(x, y float64) *Attributes {
	return a
}

func (a *Attributes) Rotate(g float64) *Attributes {
	return a
}

func (a *Attributes) Style(name, value string) *Attributes {
	a.styles[name] = value
	return a
}

func (a *Attributes) attributes() []xml.Attribute {
	return []xml.Attribute{
		xml.NewAttribute(xml.NewName("id"), a.id),
		xml.NewAttribute(xml.NewName("class"), strings.Join(a.class, " ")),
	}
}

var (
	ErrNegative = errors.New("negative value")
	ErrRange    = errors.New("out of range")
)

func negative(field string) error {
	return fmt.Errorf("%s: can not have a %w", field, ErrNegative)
}

func outOfRange(field string) error {
	return fmt.Errorf("%s: value is %w", field, ErrRange)
}
