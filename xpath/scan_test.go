package xpath

import (
	"strings"
	"testing"
)

type ScanTestCase struct {
	Query string
	Want  []Token
}

func createToken(kind rune, literal string) Token {
	return Token{
		Type:    kind,
		Literal: literal,
	}
}

func TestScanner(t *testing.T) {
	t.Run("names-and-paths", testScanNamePath)
	t.Run("namespaces-and-attributes", testScanNamespacesAndAttributes)
	t.Run("variables-and-literals", testScanVariablesAndLiterals)
	t.Run("predicates", testScanPredicates)
	t.Run("operators", testScanOperators)
	t.Run("queries", testScanFullQueries)
}

func testScanFullQueries(t *testing.T) {
	tests := []ScanTestCase{
		{
			Query: "count(/foo[/bar and @test >= 10])",
			Want: []Token{
				createToken(Name, "count"),
				createToken(BegGrp, ""),
				createToken(Slash, ""),
				createToken(Name, "foo"),
				createToken(BegPred, ""),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Name, "and"),
				createToken(Attr, ""),
				createToken(Name, "test"),
				createToken(Ge, ""),
				createToken(Number, "10"),
				createToken(EndPred, ""),
				createToken(EndGrp, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "//foo[contains(@id, 'bar')]/bar",
			Want: []Token{
				createToken(Deep, ""),
				createToken(Name, "foo"),
				createToken(BegPred, ""),
				createToken(Name, "contains"),
				createToken(BegGrp, ""),
				createToken(Attr, ""),
				createToken(Name, "id"),
				createToken(Comma, ""),
				createToken(Literal, "bar"),
				createToken(EndGrp, ""),
				createToken(EndPred, ""),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "/foo/*:bar[fn:round(test) >= 10]",
			Want: []Token{
				createToken(Slash, ""),
				createToken(Name, "foo"),
				createToken(Slash, ""),
				createToken(Mul, ""),
				createToken(Namespace, ""),
				createToken(Name, "bar"),
				createToken(BegPred, ""),
				createToken(Name, "fn"),
				createToken(Namespace, ""),
				createToken(Name, "round"),
				createToken(BegGrp, ""),
				createToken(Name, "test"),
				createToken(EndGrp, ""),
				createToken(Ge, ""),
				createToken(Number, "10"),
				createToken(EndPred, ""),
				createToken(Eof, ""),
			},
		},
	}
	testScanner(t, tests)
}

func testScanOperators(t *testing.T) {
	tests := []ScanTestCase{
		{
			Query: "1+1",
			Want: []Token{
				createToken(Number, "1"),
				createToken(Add, ""),
				createToken(Number, "1"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "1--1",
			Want: []Token{
				createToken(Number, "1"),
				createToken(Sub, ""),
				createToken(Sub, ""),
				createToken(Number, "1"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "1-1",
			Want: []Token{
				createToken(Number, "1"),
				createToken(Sub, ""),
				createToken(Number, "1"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "1 div 10",
			Want: []Token{
				createToken(Number, "1"),
				createToken(Name, "div"),
				createToken(Number, "10"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "\"foo\" || /bar",
			Want: []Token{
				createToken(Literal, "foo"),
				createToken(Concat, ""),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "$foo = 10",
			Want: []Token{
				createToken(Variable, "foo"),
				createToken(Eq, ""),
				createToken(Number, "10"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo != 10",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Ne, ""),
				createToken(Number, "10"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo < bar",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Lt, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo <= bar",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Le, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "bar >= foo",
			Want: []Token{
				createToken(Name, "bar"),
				createToken(Ge, ""),
				createToken(Name, "foo"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "bar > foo",
			Want: []Token{
				createToken(Name, "bar"),
				createToken(Gt, ""),
				createToken(Name, "foo"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo and bar",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Name, "and"),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo or bar",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Name, "or"),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
	}
	testScanner(t, tests)
}

func testScanPredicates(t *testing.T) {
	tests := []ScanTestCase{
		{
			Query: "foo[1]",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(BegPred, ""),
				createToken(Number, "1"),
				createToken(EndPred, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo[@id]",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(BegPred, ""),
				createToken(Attr, ""),
				createToken(Name, "id"),
				createToken(EndPred, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo[position() = 1]",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(BegPred, ""),
				createToken(Name, "position"),
				createToken(BegGrp, ""),
				createToken(EndGrp, ""),
				createToken(Eq, ""),
				createToken(Number, "1"),
				createToken(EndPred, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "(foo/bar)",
			Want: []Token{
				createToken(BegGrp, ""),
				createToken(Name, "foo"),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(EndGrp, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "(foo/bar)[1]",
			Want: []Token{
				createToken(BegGrp, ""),
				createToken(Name, "foo"),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(EndGrp, ""),
				createToken(BegPred, ""),
				createToken(Number, "1"),
				createToken(EndPred, ""),
				createToken(Eof, ""),
			},
		},
	}
	testScanner(t, tests)
}

func testScanVariablesAndLiterals(t *testing.T) {
	tests := []ScanTestCase{
		{
			Query: "$foobar",
			Want: []Token{
				createToken(Variable, "foobar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "$",
			Want: []Token{
				createToken(Invalid, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "'foobar'",
			Want: []Token{
				createToken(Literal, "foobar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "'foobar",
			Want: []Token{
				createToken(Invalid, "foobar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "\"foobar",
			Want: []Token{
				createToken(Invalid, "foobar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "'it''s a hello world'",
			Want: []Token{
				createToken(Literal, "it's a hello world"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "\"it's a \"\"hello world\"\"\"",
			Want: []Token{
				createToken(Literal, "it's a \"hello world\""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "\"foobar\"",
			Want: []Token{
				createToken(Literal, "foobar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "1234",
			Want: []Token{
				createToken(Number, "1234"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "1234.56",
			Want: []Token{
				createToken(Number, "1234.56"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "-1234",
			Want: []Token{
				createToken(Sub, ""),
				createToken(Number, "1234"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "0.1234",
			Want: []Token{
				createToken(Number, "0.1234"),
				createToken(Eof, ""),
			},
		},
		{
			Query: ".1",
			Want: []Token{
				createToken(Number, ".1"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "1.",
			Want: []Token{
				createToken(Number, "1."),
				createToken(Eof, ""),
			},
		},
	}
	testScanner(t, tests)
}

func testScanNamespacesAndAttributes(t *testing.T) {
	tests := []ScanTestCase{
		{
			Query: "foo:bar",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Namespace, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo:*",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Namespace, ""),
				createToken(Mul, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "*:bar",
			Want: []Token{
				createToken(Mul, ""),
				createToken(Namespace, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "*:*",
			Want: []Token{
				createToken(Mul, ""),
				createToken(Namespace, ""),
				createToken(Mul, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "@foo",
			Want: []Token{
				createToken(Attr, ""),
				createToken(Name, "foo"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "@foo:bar",
			Want: []Token{
				createToken(Attr, ""),
				createToken(Name, "foo"),
				createToken(Namespace, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "@*",
			Want: []Token{
				createToken(Attr, ""),
				createToken(Mul, ""),
				createToken(Eof, ""),
			},
		},
	}
	testScanner(t, tests)
}

func testScanNamePath(t *testing.T) {
	tests := []ScanTestCase{
		{
			Query: "foo",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo123",
			Want: []Token{
				createToken(Name, "foo123"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo_bar",
			Want: []Token{
				createToken(Name, "foo_bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "foo/bar",
			Want: []Token{
				createToken(Name, "foo"),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: ".",
			Want: []Token{
				createToken(Current, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "..",
			Want: []Token{
				createToken(Parent, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "//foo",
			Want: []Token{
				createToken(Deep, ""),
				createToken(Name, "foo"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "*",
			Want: []Token{
				createToken(Mul, ""),
				createToken(Eof, ""),
			},
		},
		{
			Query: "/foo | /bar",
			Want: []Token{
				createToken(Slash, ""),
				createToken(Name, "foo"),
				createToken(Pipe, ""),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "/foo union /bar",
			Want: []Token{
				createToken(Slash, ""),
				createToken(Name, "foo"),
				createToken(Name, "union"),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "/foo intersect /bar",
			Want: []Token{
				createToken(Slash, ""),
				createToken(Name, "foo"),
				createToken(Name, "intersect"),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "/foo except /bar",
			Want: []Token{
				createToken(Slash, ""),
				createToken(Name, "foo"),
				createToken(Name, "except"),
				createToken(Slash, ""),
				createToken(Name, "bar"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "/attribute::foo",
			Want: []Token{
				createToken(Slash, ""),
				createToken(Name, "attribute"),
				createToken(AxisSep, ""),
				createToken(Name, "foo"),
				createToken(Eof, ""),
			},
		},
		{
			Query: "/",
			Want: []Token{
				createToken(Slash, ""),
				createToken(Eof, ""),
			},
		},
	}
	testScanner(t, tests)
}

func testScanner(t *testing.T, tests []ScanTestCase) {
	t.Helper()
	for _, c := range tests {
		var (
			scan    = Scan(strings.NewReader(c.Query))
			ptr     int
			scanned int
		)
		for {
			if err := scan.Err(); err != nil {
				t.Errorf("%s(%s): error while scanning: %s", t.Name(), c.Query, err)
				break
			}
			tok := scan.Scan()
			scanned++
			if ptr >= len(c.Want) {
				t.Errorf("%s(%s): too many tokens scanned!", t.Name(), c.Query)
				break
			}
			if tok.Type != c.Want[ptr].Type || tok.Literal != c.Want[ptr].Literal {
				t.Errorf("%s(%s): token mismatched! want %s, got %s", t.Name(), c.Query, c.Want[ptr], tok)
			}
			ptr++
			if tok.Type == Eof {
				break
			}
		}
		if ptr+1 < len(c.Want) {
			t.Errorf("%s(%s): not enough tokens scanned! want %d, got %d", t.Name(), c.Query, len(c.Want), scanned)
		}
	}
}
