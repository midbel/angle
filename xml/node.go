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

func (n Name) Equal(other Name) bool {
	return n.Local == other.Local && n.Namespace.Equal(other.Namespace)
}

func (n Name) LexicalName() string {
	if n.Prefix == "" && n.URI == "" {
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

func (Document) Type() NodeType {
	return DocumentNode
}

type Element struct {
	Name
	NS         []Namespace
	Attributes []Attribute
	Children   []Node
}

func (Element) Type() NodeType {
	return ElementNode
}

type Attribute struct {
	Name
	Value string
}

type Text struct {
	Value string
}

func (Text) Type() NodeType {
	return TextNode
}

type Comment struct {
	Value string
}

func (Comment) Type() NodeType {
	return CommentNode
}

type PI struct {
	Name
	Data string
}

func (PI) Type() NodeType {
	return PINode
}
