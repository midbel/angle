package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/midbel/angle/xpath"
)

func main() {
	flag.Parse()

	r := strings.NewReader(flag.Arg(0))
	scan := xpath.Scan(r)
	for {
		tok := scan.Scan()
		fmt.Println(tok)
		if tok.Type == xpath.Eof || tok.Type == xpath.Invalid {
			break
		}
	}
}
