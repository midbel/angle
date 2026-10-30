package xpath

import "fmt"

type Token struct {
	Literal string
	Type    rune
	Position
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	case Eof:
		return "<eof>"
	case Name:
		prefix = "name"
	case Variable:
		prefix = "variable"
	case Current:
		return "<current>"
	case Parent:
		return "<parent>"
	case Attr:
		return "<attribute>"
	case Literal:
		prefix = "literal"
	case Number:
		prefix = "number"
	case Slash:
		return "<path>"
	case Deep:
		return "<deep-path>"
	case Pipe:
		return "<pipe>"
	case Comma:
		return "<comma>"
	case Namespace:
		return "<namespace>"
	case AxisSep:
		return "<axis>"
	case BegPred:
		return "<beg-predicate>"
	case EndPred:
		return "<end-predicate>"
	case BegGrp:
		return "<beg-group>"
	case EndGrp:
		return "<end-group>"
	case Eq:
		return "<equal>"
	case Ne:
		return "<notequal>"
	case Le:
		return "<lesser-equal>"
	case Lt:
		return "<lesser-than>"
	case Ge:
		return "<greater-equal>"
	case Gt:
		return "<greater-than>"
	case Add:
		return "<add>"
	case Sub:
		return "<subtract>"
	case Mul:
		return "<multiply>"
	case Div:
		return "<divide>"
	case Mod:
		return "<modulo>"
	case Concat:
		return "<concat>"
	case And:
		return "<and>"
	case Or:
		return "<or>"
	case Invalid:
		return "<invalid>"
	default:
		return "<unknown>"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}

const (
	Eof rune = iota
	Name
	Variable
	Current
	Parent
	Attr
	Literal
	Number
	Namespace
	Slash
	Deep
	Pipe
	Comma
	AxisSep
	BegPred
	EndPred
	BegGrp
	EndGrp
	Concat
	Add
	Sub
	Mul
	Div
	Mod
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
	And
	Or
	Invalid
)

const (
	powLowest int = iota
	powOr
	powAnd
	powEq
	powCmp
	powConcat
	powAdd
	powMul
	powCall
	powPrefix
)

var bindings = map[rune]int{
	BegGrp: powCall,
	Concat: powConcat,
	Add:    powAdd,
	Sub:    powAdd,
	Mul:    powMul,
	Div:    powMul,
	Mod:    powMul,
	Eq:     powEq,
	Ne:     powEq,
	Lt:     powCmp,
	Le:     powCmp,
	Gt:     powCmp,
	Ge:     powCmp,
	And:    powAnd,
	Or:     powOr,
}
