package xml

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

type Name struct {
	Prefix string
	Local  string
}

type Document struct {
	Children []Node
}

type Element struct {
	Name       Name
	Attributes []Attribute
	Children   []Node
}

type Attribute struct {
	Name  Name
	Value string
}

type Text struct {
	Value string
}

type Comment struct {
	Value string
}

type PI struct {
	Target string
	Data   string
}
