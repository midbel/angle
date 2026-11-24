package xpath

import (
	"strings"
	"testing"
)

func TestParserValidExpr(t *testing.T) {
	tests := []string{
		"/",
		"/foo",
		"/foo[1]",
		"/foo[position() = 1]",
		"/foo[@id and @active]",
		"/foo[@id or @name]",
		"/foo/bar",
		"/foo/self::bar",
		"//foo/bar",
		"/ns:foo",
		"count(/foo)",
		"/foo | /bar",
		"/foo union /bar union /quz",
		"/foo except /quz",
		"/foo[@bar = $var]",
		"/foo/(./bar | ./quz)/it[@attr = '10']",
	}
	for _, q := range tests {
		_, err := parseExpr(q)
		if err != nil {
			t.Errorf("%s(%s): unexpected error parsing expression: %s", t.Name(), q, err)
		}
	}
}

type AstTestCase struct {
	Query string
	Want  Expr
}

func TestParserAst(t *testing.T) {
	t.Run("associativity", testParserAssociativity)
	t.Run("set", testParserSet)
}

func testParserSet(t *testing.T) {
	tests := []AstTestCase{
		{
			Query: "foo | bar union quz",
			Want: createUnion(
				createUnion(
					axisStep(ChildAxis, createName("", "foo")),
					axisStep(ChildAxis, createName("", "bar")),
				),
				axisStep(ChildAxis, createName("", "quz")),
			),
		},
		{
			Query: "foo intersect bar",
			Want: createIntersect(
				axisStep(ChildAxis, createName("", "foo")),
				axisStep(ChildAxis, createName("", "bar")),
			),
		},
		{
			Query: "foo union bar except quz",
			Want: createUnion(
				axisStep(ChildAxis, createName("", "foo")),
				createExcept(
					axisStep(ChildAxis, createName("", "bar")),
					axisStep(ChildAxis, createName("", "quz")),
				),
			),
		},
		{
			Query: "foo intersect bar union quz",
			Want: createUnion(
				createIntersect(
					axisStep(ChildAxis, createName("", "foo")),
					axisStep(ChildAxis, createName("", "bar")),
				),
				axisStep(ChildAxis, createName("", "quz")),
			),
		},
		{
			Query: "foo except bar intersect quz",
			Want: createIntersect(
				createExcept(
					axisStep(ChildAxis, createName("", "foo")),
					axisStep(ChildAxis, createName("", "bar")),
				),
				axisStep(ChildAxis, createName("", "quz")),
			),
		},
	}
	testAstParser(t, tests)
}

func testParserAssociativity(t *testing.T) {
	tests := []AstTestCase{
		{
			Query: "1+2*3",
			Want: createBinary(
				createNumber(1),
				createBinary(
					createNumber(2),
					createNumber(3),
					Mul,
				),
				Add,
			),
		},
		{
			Query: "1+(2*3)",
			Want: createBinary(
				createNumber(1),
				createBinary(
					createNumber(2),
					createNumber(3),
					Mul,
				),
				Add,
			),
		},
		{
			Query: "(1+2)*3",
			Want: createBinary(
				createBinary(
					createNumber(1),
					createNumber(2),
					Add,
				),
				createNumber(3),
				Mul,
			),
		},
		{
			Query: "-1 + 2",
			Want: createBinary(
				createUnary(
					createNumber(1),
					Sub,
				),
				createNumber(2),
				Add,
			),
		},
		{
			Query: "-(1 + 2)",
			Want: createUnary(
				createBinary(
					createNumber(1),
					createNumber(2),
					Add,
				),
				Sub,
			),
		},
		{
			Query: "10 div 5 * 2",
			Want: createBinary(
				createBinary(
					createNumber(10),
					createNumber(5),
					Div,
				),
				createNumber(2),
				Mul,
			),
		},
	}
	testAstParser(t, tests)
}

func testAstParser(t *testing.T, tests []AstTestCase) {
	t.Helper()
	for _, tt := range tests {
		got, err := parseExpr(tt.Query)
		if err != nil {
			t.Errorf("%s(%s): fail to parse expression: %s", t.Name(), tt.Query, err)
			continue
		}
		if tt.Want.String() != got.String() {
			t.Errorf("%s(%s): expressions mismatched! want %s, got %s", t.Name(), tt.Query, tt.Want, got)
		}
	}
}

func parseExpr(str string) (Expr, error) {
	p := NewParser(strings.NewReader(str))
	return p.Parse()
}
