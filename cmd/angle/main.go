package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

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

func prepare() *cli.CommandTrie {
	root := cli.New()
	root.Register(single("tokenize"), &scanCmd)
	return root
}

func single(str string) []string {
	return []string{str}
}
