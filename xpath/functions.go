package xpath

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/midbel/angle/xml"
)

var (
	ErrSingleton = errors.New("singleton sequence expected")
	ErrArgument  = errors.New("invalid number of arguments given")
	ErrContext   = errors.New("missing context")
	ErrOverflow  = errors.New("overflow")
	ErrType      = errors.New("invalid type")
)

type Arguments []Sequence

func (a Arguments) Len() int {
	return len(a)
}

func (a Arguments) Sequence(i int) Sequence {
	if i >= len(a) || i < 0 {
		return nil
	}
	return a[i]
}

func (a Arguments) Node(i int) xml.Node {
	v := a.First(i)
	if v == nil {
		return nil
	}
	if v.Atomic() {
		return nil
	}
	return v.Node()
}

func (a Arguments) Atomic(i int) Item {
	v := a.First(i)
	if v == nil {
		return nil
	}
	if !v.Atomic() {
		return nil
	}
	return v
}

func (a Arguments) First(i int) Item {
	if i >= len(a) || i < 0 {
		return nil
	}
	return a[i].First()
}

func (a Arguments) Last(i int) Item {
	if i >= len(a) || i < 0 {
		return nil
	}
	return a[i].Last()
}

type BuiltinFunc func(*Context, Focus, Arguments) (Sequence, error)

type Builtin struct {
	Name       string
	UseCtxItem bool
	MinArgs    int
	MaxArgs    int
	Func       BuiltinFunc
}

func funcName(name string) qname {
	return createQName(name, nsPrefixFunc)
}

var builtins = map[qname]Builtin{}

func init() {
	registerBuiltin("string", 0, 1, true, callString)
	registerBuiltin("concat", 2, -1, false, callConcat)
	registerBuiltin("compare", 2, 3, false, callCompare)
	registerBuiltin("contains", 2, 2, false, callContains)
	registerBuiltin("substring", 2, 3, false, callSubstring)
	registerBuiltin("matches", 2, 3, false, callMatches)
	registerBuiltin("replace", 3, 4, false, callReplace)
	registerBuiltin("tokenize", 2, 3, false, callTokenize)
	registerBuiltin("substring-before", 2, 2, false, callSubstringBefore)
	registerBuiltin("substring-after", 2, 2, false, callSubstringAfter)
	registerBuiltin("normalize-space", 0, 1, true, callNormalizeSpace)
	registerBuiltin("upper-case", 1, 1, false, callUppercase)
	registerBuiltin("lower-case", 1, 1, false, callLowercase)
	registerBuiltin("starts-with", 2, 2, false, callStartsWith)
	registerBuiltin("ends-with", 2, 2, false, callEndsWith)
	registerBuiltin("string-join", 1, 2, false, callStringJoin)
	registerBuiltin("string-length", 0, 1, true, callStringLength)
	registerBuiltin("number", 0, 1, true, callNumber)
	registerBuiltin("abs", 1, 1, false, callAbs)
	registerBuiltin("floor", 1, 1, false, callFloor)
	registerBuiltin("ceiling", 1, 1, false, callCeiling)
	registerBuiltin("round", 1, 2, false, callRound)
	registerBuiltin("sum", 1, 2, false, callSum)
	registerBuiltin("avg", 1, 1, false, callAvg)
	registerBuiltin("max", 1, 1, false, callMax)
	registerBuiltin("min", 1, 1, false, callMin)
	registerBuiltin("boolean", 1, 1, false, callBoolean)
	registerBuiltin("true", 0, 0, false, callTrue)
	registerBuiltin("false", 0, 0, false, callFalse)
	registerBuiltin("not", 1, 1, false, callNot)
	registerBuiltin("count", 1, 1, false, callCount)
	registerBuiltin("position", 0, 0, false, callPosition)
	registerBuiltin("last", 0, 0, false, callLast)
	registerBuiltin("name", 0, 1, true, callName)
	registerBuiltin("local-name", 0, 1, true, callLocalName)
	registerBuiltin("namespace-uri", 0, 1, true, callNamespaceUri)
	registerBuiltin("node-name", 0, 1, true, callNodeName)
	registerBuiltin("root", 0, 1, true, callRoot)
	registerBuiltin("has-children", 0, 1, true, callHasChildren)
	registerBuiltin("exists", 1, 1, false, callExists)
	registerBuiltin("empty", 1, 1, false, callEmpty)
	registerBuiltin("tail", 1, 1, false, callTail)
	registerBuiltin("head", 1, 1, false, callHead)
	registerBuiltin("reverse", 1, 1, false, callReverse)
	registerBuiltin("zero-or-one", 1, 1, false, callZeroOrOne)
	registerBuiltin("one-or-more", 1, 1, false, callOneOrMore)
	registerBuiltin("exactly-one", 1, 1, false, callExactlyOne)
	registerBuiltin("distinct-values", 1, 2, false, callDistinctValues)
}

func callString(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := stringSeq(args.Sequence(0))
	return Singleton(str), err
}

func callCompare(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	fst, err := stringSeq(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	snd, err := stringSeq(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	res := strings.Compare(fst, snd)
	return Singleton(int64(res)), nil
}

func callConcat(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	var res string
	for i := range args.Len() {
		str, err := stringSeq(args.Sequence(i))
		if err != nil {
			return nil, err
		}
		res += str
	}
	return Singleton(res), nil
}

func callContains(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := stringSeq(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	sub, err := stringSeq(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	ok := strings.Contains(str, sub)
	return Singleton(ok), nil
}

func callSubstring(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := getStringFromSequence(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	var (
		offset float64
		res    = []rune(str)
		size   = float64(len(str))
	)
	offset, err = getFloatFromSequence(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	offset -= 1
	if seq := args.Sequence(2); !seq.Empty() {
		size, err = getFloatFromSequence(seq)
		if err != nil {
			return nil, err
		}
	}
	if offset < 0 {
		size += offset
		offset = 0
	}
	if z := len(res); int(offset+size) >= z {
		size = float64(z) - offset
	}
	res = res[int(offset) : int(offset)+int(size)]
	return Singleton(string(res)), nil
}

func callSubstringBefore(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := getStringFromSequence(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	cut, err := getStringFromSequence(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	str, _, ok := strings.Cut(str, cut)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(str), nil
}

func callSubstringAfter(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := getStringFromSequence(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	cut, err := getStringFromSequence(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	_, str, ok := strings.Cut(str, cut)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(str), nil
}

func callMatches(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := getStringFromSequence(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	match, err := getStringFromSequence(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	ok, err := regexp.MatchString(match, str)
	if err != nil {
		return nil, err
	}
	return Singleton(ok), nil
}

func callReplace(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := getStringFromSequence(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	pattern, err := getStringFromSequence(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	var replace string
	if seq := args.Sequence(2); !seq.Empty() {
		replace, err = getStringFromSequence(seq)
		if err != nil {
			return nil, err
		}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	str = re.ReplaceAllString(str, replace)
	return Singleton(str), nil
}

func callTokenize(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := getStringFromSequence(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	pattern, err := getStringFromSequence(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var seq Sequence
	for _, str := range re.Split(str, -1) {
		seq = append(seq, newAtomicItem(str))
	}
	return seq, nil
}

func callNormalizeSpace(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	var (
		res  []rune
		prev rune
	)
	str, err := stringSeq(args.Sequence(0))
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

func callUppercase(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := stringSeq(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	str = strings.ToUpper(str)
	return Singleton(str), nil
}

func callLowercase(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := stringSeq(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	str = strings.ToLower(str)
	return Singleton(str), nil
}

func callStartsWith(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := stringSeq(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	prefix, err := stringSeq(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	ok := strings.HasPrefix(str, prefix)
	return Singleton(ok), nil
}

func callEndsWith(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := stringSeq(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	suffix, err := stringSeq(args.Sequence(1))
	if err != nil {
		return nil, err
	}
	ok := strings.HasSuffix(str, suffix)
	return Singleton(ok), nil
}

func callStringJoin(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	var (
		str []string
		sep string
		err error
	)
	for _, it := range args.Sequence(0) {
		s, err := castToString(it.Value())
		if err != nil {
			return nil, err
		}
		str = append(str, s)
	}
	if seq := args.Sequence(1); !seq.Empty() {
		sep, err = getStringFromSequence(seq)
		if err != nil {
			return nil, err
		}
	}
	return Singleton(strings.Join(str, sep)), nil
}

func callStringLength(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	str, err := stringSeq(args.Sequence(0))
	if err != nil {
		return nil, err
	}
	n := utf8.RuneCountInString(str)
	return Singleton(int64(n)), nil
}

func callNumber(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Empty() {
		return Singleton(math.NaN()), nil
	}
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	val, err := atomize(seq.First().Value())
	if err != nil {
		return nil, err
	}
	res, err := castToFloat(val)
	if err != nil {
		res = math.NaN()
	}
	return Singleton(res), nil
}

func callAbs(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return numberSeq(args.Sequence(0), func(val any) (Sequence, error) {
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

func callFloor(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return numberSeq(args.Sequence(0), func(val any) (Sequence, error) {
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

func callCeiling(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return numberSeq(args.Sequence(0), func(val any) (Sequence, error) {
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

func callRound(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return numberSeq(args.Sequence(0), func(val any) (Sequence, error) {
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

func callSum(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
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

func callAvg(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
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

func callMin(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
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
		if i == 0 {
			res = x
		} else {
			res = min(res, x)
		}
	}
	return Singleton(res), nil
}

func callMax(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
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
		if i == 0 {
			res = x
		} else {
			res = max(res, x)
		}
	}
	return Singleton(res), nil
}

func callBoolean(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	ok, err := effectiveBooleanValue(args.Sequence(0))
	return Singleton(ok), err
}

func callTrue(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return Singleton(true), nil
}

func callFalse(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return Singleton(false), nil
}

func callNot(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	ok, err := effectiveBooleanValue(args.Sequence(0))
	return Singleton(!ok), err
}

func callCount(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	n := args.Sequence(0).Len()
	return Singleton(int64(n)), nil
}

func callPosition(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return Singleton(int64(curr.Index)), nil
}

func callLast(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return Singleton(int64(curr.Size)), nil
}

func callName(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq.Empty() {
		return Singleton(""), nil
	}
	node := args.Node(0)
	if node == nil {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(node)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.QualifiedName()), nil
}

func callLocalName(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq.Empty() {
		return Singleton(""), nil
	}
	node := args.Node(0)
	if node == nil {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(node)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.LocalName()), nil
}

func callNamespaceUri(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq.Empty() {
		return Singleton(""), nil
	}
	node := args.Node(0)
	if node == nil {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(node)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.URI), nil
}

func callNodeName(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq.Empty() {
		return nil, nil
	}
	node := args.Node(0)
	if node == nil {
		return nil, fmt.Errorf("node expected")
	}
	n, ok := getNameFromNode(node)
	if !ok {
		return Singleton(""), nil
	}
	return Singleton(n.ExpandedName()), nil
}

func callRoot(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq.Empty() {
		return nil, nil
	}
	node := args.Node(0)
	if node == nil {
		return nil, fmt.Errorf("node expected")
	}
	root := traverseRoot(node)
	return Singleton(root), nil
}

func callHasChildren(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	if seq.Empty() {
		return Singleton(false), nil
	}
	node := args.Node(0)
	if node == nil {
		return nil, fmt.Errorf("node expected")
	}
	children := traverseChild(node)
	return Singleton(len(children) > 0), nil
}

func callExists(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	ok := !args.Sequence(0).Empty()
	return Singleton(ok), nil
}

func callEmpty(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	ok := args.Sequence(0).Empty()
	return Singleton(ok), nil
}

func callTail(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() <= 1 {
		return nil, nil
	}
	var res Sequence
	for i := range seq[1:] {
		res = append(res, seq[i])
	}
	return res, nil
}

func callHead(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Empty() {
		return nil, nil
	}
	return Singleton(seq.First()), nil
}

func callReverse(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Empty() {
		return nil, nil
	}
	slices.Reverse(seq)
	return seq, nil
}

func callZeroOrOne(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Len() > 1 {
		return nil, fmt.Errorf("empty/singleton sequence expected")
	}
	return seq, nil
}

func callOneOrMore(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if seq.Empty() {
		return nil, fmt.Errorf("non empty sequence expected")
	}
	return seq, nil
}

func callExactlyOne(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	seq := args.Sequence(0)
	if !seq.Singleton() {
		return nil, ErrSingleton
	}
	return seq, nil
}

func callDistinctValues(ctx *Context, curr Focus, args Arguments) (Sequence, error) {
	return nil, nil
}

func registerBuiltin(name string, minArgs, maxArgs int, useCtxItem bool, do BuiltinFunc) {
	qn := funcName(name)
	builtins[qn] = Builtin{
		Name:       name,
		Func:       do,
		UseCtxItem: useCtxItem,
		MinArgs:    minArgs,
		MaxArgs:    maxArgs,
	}
}

func stringSeq(seq Sequence) (string, error) {
	if seq.Len() > 1 {
		return "", ErrSingleton
	}
	if seq.Empty() {
		return "", nil
	}
	return stringify(seq.First())
}

func numberSeq(seq Sequence, do func(val any) (Sequence, error)) (Sequence, error) {
	if seq.Empty() {
		return nil, nil
	}
	if seq.Len() > 1 {
		return nil, ErrSingleton
	}
	val, err := atomize(seq.First().Value())
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
	return castToString(seq.First().Value())
}

func getFloatFromSequence(seq Sequence) (float64, error) {
	if seq.Len() != 1 {
		return 0, ErrSingleton
	}
	val, err := atomize(seq.First().Value())
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
	return res, err
}
