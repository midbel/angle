package xml

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type NodeType uint8

const (
	DocumentNode NodeType = iota
	ElementNode
	TextNode
	CommentNode
	PINode
)

type Node interface {
	Type() NodeType
}

type Namespace struct {
	Prefix string
	URI    string
}

func (n Namespace) Zero() bool {
	return n.Prefix == "" && n.URI == ""
}

func (n Namespace) Equal(other Namespace) bool {
	return n.Prefix == other.Prefix && n.URI == other.URI
}

type Name struct {
	Local string
	Namespace
}

func IsValidName(str string) bool {
	r, size := utf8.DecodeRuneInString(str)
	if !IsValidStartNameChar(r) {
		return false
	}
	for _, r := range str[size:] {
		if !IsValidNameChar(r) {
			return false
		}
	}
	return true
}

func NewName(str string) Name {
	return Name{
		Local: str,
	}
}

func ParseName(str string) (Name, error) {
	valid := IsValidName(str)
	if !valid {
		return Name{}, ErrName
	}
	var (
		name Name
		ok   bool
	)
	name.Prefix, name.Local, ok = strings.Cut(str, ":")
	if !ok {
		name.Prefix = ""
		name.Local = str
	} else {
		if name.Prefix == "" || name.Local == "" {
			return name, ErrName
		}
	}
	return name, nil
}

func (n Name) Zero() bool {
	return n.Local == "" && n.Namespace.Zero()
}

func (n Name) Equal(other Name) bool {
	return n.Local == other.Local && n.Namespace.Equal(other.Namespace)
}

func (n Name) ExpandedName() string {
	if n.URI == "" {
		return n.Local
	}
	return fmt.Sprintf("%s:%s", n.URI, n.Local)
}

func (n Name) QualifiedName() string {
	if n.Prefix == "" {
		return n.Local
	}
	return fmt.Sprintf("%s:%s", n.Prefix, n.Local)
}

type Document struct {
	Children []Node
}

func NewDocument(root Node) *Document {
	return &Document{
		Children: []Node{root},
	}
}

func (d *Document) Root() (Node, error) {
	for _, c := range d.Children {
		el, ok := c.(*Element)
		if ok {
			return el, nil
		}
	}
	return nil, fmt.Errorf("missing root")
}

func (Document) Type() NodeType {
	return DocumentNode
}

type Element struct {
	Name
	NS         []Namespace
	Attributes []Attribute
	Children   []Node
}

func NewElement(name Name) *Element {
	return &Element{
		Name: name,
	}
}

func (e *Element) FirstChild() (Node, error) {
	n := len(e.Children)
	if n == 0 {
		return nil, fmt.Errorf("no child available")
	}
	return e.Children[0], nil
}

func (e *Element) LastChild() (Node, error) {
	n := len(e.Children)
	if n == 0 {
		return nil, fmt.Errorf("no child available")
	}
	return e.Children[n-1], nil
}

func (e *Element) ElementByName(name string) (Node, error) {
	all := e.ElementsByName(name)
	if len(all) == 0 {
		return nil, fmt.Errorf("empty set")
	}
	return all[0], nil
}

func (e *Element) ElementsByName(name string) []Node {
	var list []Node
	for _, c := range e.Children {
		el, ok := c.(*Element)
		if !ok {
			continue
		}
		if el.Local == name {
			list = append(list, el)
		}
	}
	return list
}

func (*Element) Type() NodeType {
	return ElementNode
}

type Attribute struct {
	Name
	Value string
}

func NewAttribute(name Name, value string) Attribute {
	return Attribute{
		Name:  name,
		Value: value,
	}
}

type Text struct {
	Value string
}

func NewText(text string) *Text {
	return &Text{
		Value: text,
	}
}

func (*Text) Type() NodeType {
	return TextNode
}

type Comment struct {
	Value string
}

func NewComment(text string) *Comment {
	return &Comment{
		Value: text,
	}
}

func (*Comment) Type() NodeType {
	return CommentNode
}

type PI struct {
	Name
	Data string
}

func NewPI(name Name, data string) *PI {
	return &PI{
		Name: name,
		Data: data,
	}
}

func (*PI) Type() NodeType {
	return PINode
}
