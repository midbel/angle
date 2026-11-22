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

func parseExpr(str string) (Expr, error) {
	p := NewParser(strings.NewReader(str))
	return p.Parse()
}
