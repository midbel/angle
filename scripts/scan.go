package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/angle/xpath"
)

func main() {
	lex := flag.Bool("x", false, "scan expression")
	flag.Parse()

	var (
		rs  = strings.NewReader(flag.Arg(0))
		err error
	)
	if *lex {
		err = scan(rs)
	} else {
		err = parse(rs)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parse(r io.Reader) error {
	p := xpath.NewParser(r)
	e, err := p.Parse()
	if err == nil {
		fmt.Println(e)
	}
	return err
}

func scan(r io.Reader) error {
	scan := xpath.Scan(r)
	for {
		tok := scan.Scan()
		fmt.Println(tok)
		if tok.Type == xpath.Eof || tok.Type == xpath.Invalid {
			break
		}
	}
	return nil
}
