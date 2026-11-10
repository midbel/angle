package xpath

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
	registerBuiltin("path", callPath)
}

func callString(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callConcat(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callContains(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callSubstring(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callNormalizeSpace(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callUppercase(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callLowercase(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callStartsWith(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callEndsWith(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callStringJoin(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callStringLength(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callNumber(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callAbs(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callFloor(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callCeiling(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callRound(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callSum(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callAvg(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callMin(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callMax(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callBoolean(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callTrue(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callFalse(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callNot(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callExists(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callEmpty(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callCount(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callPosition(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callLast(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callName(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callLocalName(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callNamespaceUri(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callNodeName(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callRoot(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callHasChildren(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func callPath(ctx *Context, curr Focus, args []Sequence) (Sequence, error) {
	return nil, nil
}

func registerBuiltin(name string, do BuiltinFunc) {
	qn := funcName(name)
	builtins[qn] = Builtin{
		Name: name,
		Func: do,
	}
}
