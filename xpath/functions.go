package xpath

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrSingleton = errors.New("singleton sequence expected")
	ErrArgument  = errors.New("invalid number of arguments given")
	ErrContext   = errors.New("missing context")
	ErrOverflow  = errors.New("overflow")
	ErrType      = errors.New("invalid type")
)

type BuiltinFunc func(*Context, Focus, []Sequence) (Sequence, error)

type Builtin struct {
	Name    string
	MinArgs int
	MaxArgs int
	Func    BuiltinFunc
}

func funcName(name string) qname {
	return createQName(name, nsPrefixFunc)
}

var builtins = map[qname]Builtin{}

func init() {
	registerBuiltin("string", callString)
	registerBuiltin("concat", callConcat)
	registerBuiltin("contains", callContains)
	registerBuiltin("substring", callSubstring)
	registerBuiltin("matches", callMatches)
	registerBuiltin("replace", callReplace)
	registerBuiltin("tokenize", callTokenize)
	registerBuiltin("substring-before", callSubstringBefore)
	registerBuiltin("substring-after", callSubstringAfter)
	registerBuiltin("normalize-space", callNormalizeSpace)
	registerBuiltin("upper-case", callUppercase)
	registerBuiltin("lower-case", callLowercase)
	registerBuiltin("starts-with", callStartsWith)
	registerBuiltin("ends-with", callEndsWith)
	registerBuiltin("string-join", callStringJoin)
	registerBuiltin("string-length", callStringLength)
	registerBuiltin("number", callNumber)
	registerBuiltin("abs", callAbs)
	registerBuiltin("floor", callFloor)
	registerBuiltin("ceiling", callCeiling)
	registerBuiltin("round", callRound)
	registerBuiltin("sum", callSum)
	registerBuiltin("avg", callAvg)
	registerBuiltin("max", callMax)
	registerBuiltin("min", callMin)
	registerBuiltin("boolean", callBoolean)
	registerBuiltin("true", callTrue)
	registerBuiltin("false", callFalse)
	registerBuiltin("not", callNot)
	registerBuiltin("exists", callExists)
	registerBuiltin("empty", callEmpty)
	registerBuiltin("count", callCount)
	registerBuiltin("position", callPosition)
	registerBuiltin("last", callLast)
	registerBuiltin("name", callName)
	registerBuiltin("local-name", callLocalName)
	registerBuiltin("namespace-uri", callNamespaceUri)
	registerBuiltin("node-name", callNodeName)
	registerBuiltin("root", callRoot)
	registerBuiltin("has-children", callHasChildren)
	// registerBuiltin("path", callPath)
}

func callString(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	str, err := stringSeq(seq)
	return Singleton(str), err
}

func callConcat(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	var res string
	for i := range args {
		str, err := stringSeq(args[i])
		if err != nil {
			return nil, err
		}
		res += str
	}
	return Singleton(res), nil
}

func callContains(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := stringSeq(args[0])
	if err != nil {
		return nil, err
	}
	sub, err := stringSeq(args[1])
	if err != nil {
		return nil, err
	}
	ok := strings.Contains(str, sub)
	return Singleton(ok), nil
}

func callSubstring(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := getStringFromSequence(args[0])
	if err != nil {
		return nil, err
	}
	var (
		offset float64
		size   float64 = float64(len(str))
	)
	offset, err = getFloatFromSequence(args[1])
	if err != nil {
		return nil, err
	}
	offset -= 1
	if len(args) == 3 {
		size, err = getFloatFromSequence(args[2])
		if err != nil {
			return nil, err
		}
	}
	if offset < 0 {
		size += offset
		offset = 0
	}
	if z := len(str); int(offset+size) >= z {
		size = float64(z) - offset
	}
	str = str[int(offset) : int(offset)+int(size)]
	return Singleton(str), nil
}

func callSubstringBefore(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := getStringFromSequence(args[0])
	if err != nil {
		return nil, err
	}
	cut, err := getStringFromSequence(args[1])
	if err != nil {
		return nil, err
	}
	str, _, ok := strings.Cut(str, cut)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(str), nil
}

func callSubstringAfter(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := getStringFromSequence(args[0])
	if err != nil {
		return nil, err
	}
	cut, err := getStringFromSequence(args[1])
	if err != nil {
		return nil, err
	}
	_, str, ok := strings.Cut(str, cut)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(str), nil
}

func callMatches(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callReplace(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callTokenize(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callNormalizeSpace(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	var (
		res  []rune
		prev rune
		seq  = args[0]
	)
	str, err := stringSeq(seq)
	if err != nil {
		return nil, err
	}
	str = strings.TrimSpace(str)
	res = make([]rune, 0, len(str))
	for i := 0; i < len(str); {
		c, z := utf8.DecodeRuneInString(str[i:])
		i += z
		if c == ' ' && prev == c {
			continue
		}
		res = append(res, c)
		prev = c
	}
	return Singleton(string(res)), nil
}

func callUppercase(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := stringSeq(args[0])
	if err != nil {
		return nil, err
	}
	str = strings.ToUpper(str)
	return Singleton(str), nil
}

func callLowercase(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := stringSeq(args[0])
	if err != nil {
		return nil, err
	}
	str = strings.ToLower(str)
	return Singleton(str), nil
}

func callStartsWith(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := stringSeq(args[0])
	if err != nil {
		return nil, err
	}
	prefix, err := stringSeq(args[1])
	if err != nil {
		return nil, err
	}
	ok := strings.HasPrefix(str, prefix)
	return Singleton(ok), nil
}

func callEndsWith(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	str, err := stringSeq(args[0])
	if err != nil {
		return nil, err
	}
	suffix, err := stringSeq(args[1])
	if err != nil {
		return nil, err
	}
	ok := strings.HasSuffix(str, suffix)
	return Singleton(ok), nil
}

func callStringJoin(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callStringLength(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	str, err := stringSeq(seq)
	if err != nil {
		return nil, err
	}
	n := utf8.RuneCountInString(str)
	return Singleton(int64(n)), nil
}

func callNumber(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	if seq.Empty() {
		return Singleton(math.NaN()), nil
	}
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	val, err := atomize(seq[0].Value())
	if err != nil {
		return nil, err
	}
	res, err := castToFloat(val)
	if err != nil {
		res = math.NaN()
	}
	return Singleton(res), nil
}

func callAbs(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return numberSeq(args[0], func(val any) (Sequence, error) {
		switch v := val.(type) {
		case int64:
			if v == math.MinInt64 {
				return nil, ErrOverflow
			}
			if v < 0 {
				v = -v
			}
			return Singleton(v), nil
		case float64:
			return Singleton(math.Abs(v)), nil
		default:
			return nil, ErrType
		}
	})
}

func callFloor(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return numberSeq(args[0], func(val any) (Sequence, error) {
		switch v := val.(type) {
		case int64:
			return Singleton(v), nil
		case float64:
			return Singleton(math.Floor(v)), nil
		default:
			return nil, ErrType
		}
	})
}

func callCeiling(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return numberSeq(args[0], func(val any) (Sequence, error) {
		switch v := val.(type) {
		case int64:
			return Singleton(v), nil
		case float64:
			return Singleton(math.Ceil(v)), nil
		default:
			return nil, ErrType
		}
	})
}

func callRound(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return numberSeq(args[0], func(val any) (Sequence, error) {
		switch v := val.(type) {
		case int64:
			return Singleton(v), nil
		case float64:
			if math.IsNaN(v) || math.IsInf(v, 0) || v == 0 {
				return Singleton(v), nil
			}
			return Singleton(math.Round(v)), nil
		default:
			return nil, ErrType
		}
	})
}

func callSum(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	seq := args[0]
	if seq.Empty() {
		if len(args) > 1 {
			return args[1], nil
		}
		return Singleton(0.0), nil
	}
	var sum float64
	for i := range seq {
		v, err := atomize(seq[i].Value())
		if err != nil {
			return nil, err
		}
		x, err := castToFloat(v)
		if err != nil {
			return nil, err
		}
		sum += x
	}
	return Singleton(sum), nil
}

func callAvg(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	seq := args[0]
	if seq.Empty() {
		return nil, nil
	}
	var sum float64
	for i := range seq {
		v, err := atomize(seq[i].Value())
		if err != nil {
			return nil, err
		}
		x, err := castToFloat(v)
		if err != nil {
			return nil, err
		}
		sum += x
	}
	return Singleton(sum / float64(seq.Len())), nil
}

func callMin(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	seq := args[0]
	if seq.Empty() {
		return nil, nil
	}
	var res float64
	for i := range seq {
		v, err := atomize(seq[i].Value())
		if err != nil {
			return nil, err
		}
		x, err := castToFloat(v)
		if err != nil {
			return nil, err
		}
		res = min(res, x)
	}
	return Singleton(res), nil
}

func callMax(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	seq := args[0]
	if seq.Empty() {
		return nil, nil
	}
	var res float64
	for i := range seq {
		v, err := atomize(seq[i].Value())
		if err != nil {
			return nil, err
		}
		x, err := castToFloat(v)
		if err != nil {
			return nil, err
		}
		res = max(res, x)
	}
	return Singleton(res), nil
}

func callBoolean(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	ok, err := effectiveBooleanValue(args[0])
	return Singleton(ok), err
}

func callTrue(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return Singleton(true), nil
}

func callFalse(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return Singleton(false), nil
}

func callNot(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	ok, err := effectiveBooleanValue(args[0])
	return Singleton(!ok), err
}

func callExists(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	ok := !args[0].Empty()
	return Singleton(ok), nil
}

func callEmpty(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	ok := args[0].Empty()
	return Singleton(ok), nil
}

func callCount(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	n := args[0].Len()
	return Singleton(int64(n)), nil
}

func callPosition(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return Singleton(int64(curr.Index)), nil
}

func callLast(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return Singleton(int64(curr.Size)), nil
}

func callName(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq[0].Atomic() {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(seq[0].Node())
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.QualifiedName()), nil
}

func callLocalName(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq[0].Atomic() {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(seq[0].Node())
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.LocalName()), nil
}

func callNamespaceUri(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq[0].Atomic() {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(seq[0].Node())
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.URI), nil
}

func callNodeName(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq[0].Atomic() {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(seq[0].Node())
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.ExpandedName()), nil
}

func callRoot(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq[0].Atomic() {
		return nil, fmt.Errorf("node expected")
	}
	root := traverseRoot(seq[0].Node())
	return Singleton(root), nil
}

func callHasChildren(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	if len(args) == 0 {
		if curr.Item == nil {
			return nil, ErrContext
		}
		args = append(args, Singleton(curr.Item))
	}
	seq := args[0]
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq[0].Atomic() {
		return nil, fmt.Errorf("node expected")
	}
	children := traverseChild(seq[0].Node())
	return Singleton(len(children) > 0), nil
}

// func callPath(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
// 	return nil, nil
// }

func registerBuiltin(name string, do BuiltinFunc) {
	qn := funcName(name)
	builtins[qn] = Builtin{
		Name:    name,
		Func:    do,
		MinArgs: -1,
		MaxArgs: -1,
	}
}

func stringSeq(seq Sequence) (string, error) {
	if seq.Len() > 1 {
		return "", ErrSingleton
	}
	if seq.Empty() {
		return "", nil
	}
	return stringify(seq[0])
}

func numberSeq(seq Sequence, do func(val any) (Sequence, error)) (Sequence, error) {
	if seq.Empty() {
		return nil, nil
	}
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	val, err := atomize(seq[0].Value())
	if err != nil {
		return nil, err
	}
	val, ok := castToNumber(val)
	if !ok {
		return nil, ErrType
	}
	return do(val)
}

func castToString(val any) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case untyped:
		return string(v), nil
	default:
		return "", ErrType
	}
}

func castToNumber(val any) (any, bool) {
	switch v := val.(type) {
	case int64, float64:
		return val, true
	case untyped:
		f, err := strconv.ParseFloat(string(v), 64)
		if err == nil {
			return f, true
		}
		return nil, false
	default:
		return nil, false
	}
}

func getStringFromSequence(seq Sequence) (string, error) {
	if seq.Empty() {
		return "", nil
	}
	if seq.Len() > 1 {
		return "", ErrSingleton
	}
	return castToString(seq[0].Value())
}

func getFloatFromSequence(seq Sequence) (float64, error) {
	if seq.Len() != 1 {
		return 0, ErrSingleton
	}
	val, err := atomize(seq[0].Value())
	if err != nil {
		return 0, err
	}
	return castToFloat(val)
}

func castToFloat(val any) (float64, error) {
	var (
		res float64
		err error
	)
	switch v := val.(type) {
	case float64:
		res = v
	case int64:
		res = float64(v)
	case string:
		res, err = strconv.ParseFloat(v, 64)
		if err != nil {
			err = ErrConvert
		}
	case untyped:
		return castToFloat(string(v))
	default:
		res = math.NaN()
	}
	return res, nil
}
