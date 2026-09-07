package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/angle/xml"
	"github.com/midbel/cli"
)

var errFail = errors.New("fail")

func main() {
	var (
		set  = cli.NewFlagSet("angle")
		root = prepare()
	)
	if err := set.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			root.Help()
			os.Exit(2)
		}
	}
	err := root.Execute(set.Args())
	if err != nil {
		if s, ok := err.(cli.SuggestionError); ok && len(s.Others) > 0 {
			fmt.Fprintln(os.Stderr, "similar command(s)")
			for _, n := range s.Others {
				fmt.Fprintln(os.Stderr, "-", n)
			}
		}
		if !errors.Is(err, errFail) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

var scanCmd = cli.Command{
	Name:    "tokenize",
	Alias:   []string{"scan", "lex"},
	Summary: "",
	Usage:   "tokenize <file>",
	Handler: &scanCommand{},
}

type scanCommand struct{}

func (c scanCommand) Run(args []string) error {
	set := cli.NewFlagSet("tokenize")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		cli.FailIO(err)
	}
	defer r.Close()

	for tok, err := range xml.Scan(r) {
		fmt.Fprintf(cli.Stdout, "%-6s | %12s | %s", tok.Position, tok.Type, tok.Literal)
		fmt.Fprintln(cli.Stdout)
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintln(cli.Stderr, err)
			return err
		}
		if tok.Type == xml.TokInvalid || tok.Type == xml.TokEof {
			break
		}
	}
	return nil
}

var echoCmd = cli.Command{
	Name:    "echo",
	Alias:   []string{"debug", "print"},
	Summary: "",
	Usage:   "echo <file>",
	Handler: &echoCommand{},
}

type echoCommand struct{}

func (c echoCommand) Run(args []string) error {
	set := cli.NewFlagSet("echo")
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		cli.FailIO(err)
	}
	defer r.Close()
	return Echo(r)
}

var fmtCmd = cli.Command{
	Name:    "format",
	Alias:   []string{"fmt", "rewrite"},
	Summary: "",
	Usage:   "format [-f <output>] <input>",
	Handler: &fmtCommand{},
}

type fmtCommand struct {
	Output io.WriteCloser
	Compact bool
}

func (c fmtCommand) Run(args []string) error {
	set := cli.NewFlagSet("format")
	set.BoolVar(&c.Compact, "c", false, "compact")
	set.Func("f", "", func(str string) error {
		w, err := os.Create(str)
		if err == nil {
			c.Output = w
		}
		return err
	})
	if err := set.Parse(args); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return cli.ErrUsage
	}
	r, err := os.Open(set.Arg(0))
	if err != nil {
		cli.FailIO(err)
	}
	defer r.Close()
	if c.Output == nil {
		c.Output = cli.Stdout
	} else {
		defer c.Output.Close()
	}
	f := xml.NewFormatter(c.Output, r)
	f.SetCompact(c.Compact)
	return f.Format()
}

func prepare() *cli.CommandTrie {
	root := cli.New()
	root.Register(single("tokenize"), &scanCmd)
	root.Register(single("echo"), &echoCmd)
	root.Register(single("format"), &fmtCmd)
	return root
}

func single(str string) []string {
	return []string{str}
}

type echoHandler struct {
	depth int
	ws    *bufio.Writer
}

func Echo(r io.Reader) error {
	h := echoHandler{
		ws: bufio.NewWriter(cli.Stdout),
	}
	defer h.done()

	rs := xml.NewReader(r)
	return rs.Read(&h)
}

func (h *echoHandler) done() {
	h.ws.Flush()
}

func (h *echoHandler) enter() {
	h.depth++
}

func (h *echoHandler) leave() {
	h.depth--
}

func (h *echoHandler) writeIndent() {
	for range h.depth {
		h.writeString(" ")
	}
}

func (h *echoHandler) writeString(str string) {
	h.ws.WriteString(str)
}

func (h *echoHandler) writeBlank() {
	h.ws.WriteRune(' ')
}

func (h *echoHandler) writeNL() {
	h.ws.WriteRune('\n')
}

func (h *echoHandler) OnStartElement(el xml.Element) error {
	defer h.enter()
	h.writeIndent()
	h.writeString("START")
	h.writeBlank()
	h.writeString(el.LexicalName())
	h.writeNL()
	for _, a := range el.Attributes {
		h.writeIndent()
		h.writeBlank()
		h.writeString("-")
		h.writeBlank()
		h.writeString("@" + a.LexicalName())
		h.writeString("=")
		h.writeString(a.Value)
		h.writeNL()
	}
	return nil
}

func (h *echoHandler) OnCloseElement(n xml.Name) error {
	h.leave()
	h.writeIndent()
	h.writeString("CLOSE")
	h.writeBlank()
	h.writeString(n.QualifiedName())
	h.writeNL()
	return nil
}

func (h *echoHandler) OnText(t xml.Text) error {
	h.writeIndent()
	h.writeString("TEXT")
	h.writeBlank()
	h.writeString(strings.TrimSpace(t.Value))
	h.writeNL()
	return nil
}

func (h *echoHandler) OnComment(c xml.Comment) error {
	h.writeIndent()
	h.writeString("COMMENT")
	h.writeBlank()
	h.writeString(strings.TrimSpace(c.Value))
	h.writeNL()
	return nil
}

func (h *echoHandler) OnPI(p xml.PI) error {
	h.writeIndent()
	h.writeString("PI")
	h.writeBlank()
	h.writeString(p.QualifiedName())
	h.writeNL()
	return nil
}
