package main

import (
	"os"
	"slices"
	"strconv"

	"github.com/midbel/angle/dispatch"
	"github.com/midbel/angle/xml"
	"github.com/midbel/angle/xpath"
)

func createContext(file string) (*xpath.Context, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var (
		ch = newContextHandler()
		rs = dispatch.New()
	)
	rs.OnOpen(xml.NewName("namespace"), ch.onNamespace)
	rs.OnOpen(xml.NewName("variable"), ch.onVariable)
	if err := rs.Read(r); err != nil {
		return nil, err
	}
	return ch.ctx, nil
}

type contextHandler struct {
	ctx *xpath.Context
}

func newContextHandler() *contextHandler {
	return &contextHandler{
		ctx: xpath.NewContext(),
	}
}

func (c *contextHandler) onNamespace(r *dispatch.Reader, el xml.Element) (dispatch.Result, error) {
	ix := slices.IndexFunc(el.Attributes, func(a *xml.Attribute) bool {
		return a.Local == "prefix"
	})
	if ix < 0 {
		return dispatch.Done, nil
	}
	prefix := el.Attributes[ix].Value
	r.OnTextOnce(func(r *dispatch.Reader, txt xml.Text) (dispatch.Result, error) {
		c.ctx.RegisterNS(prefix, txt.Content())
		return dispatch.Done, nil
	})
	return dispatch.Done, nil
}

func (c *contextHandler) onVariable(r *dispatch.Reader, el xml.Element) (dispatch.Result, error) {
	ix := slices.IndexFunc(el.Attributes, func(a *xml.Attribute) bool {
		return a.Local == "ident"
	})
	if ix < 0 {
		return dispatch.Done, nil
	}
	ident := el.Attributes[ix].Value

	ix = slices.IndexFunc(el.Attributes, func(a *xml.Attribute) bool {
		return a.Local == "type"
	})
	var kind string
	if ix >= 0 {
		kind = el.Attributes[ix].Value
	}
	r.OnTextOnce(func(r *dispatch.Reader, txt xml.Text) (dispatch.Result, error) {
		var (
			value any
			err   error
		)
		switch str := txt.Content(); kind {
		case "float", "number":
			value, err = strconv.ParseFloat(str, 64)
		case "int", "integer":
			value, err = strconv.ParseInt(str, 0, 64)
		case "bool", "boolean":
			value, err = strconv.ParseBool(str)
		default:
			value = str
		}
		if err == nil {
			c.ctx.RegisterVar(ident, value)
		}
		return dispatch.Done, err
	})
	return dispatch.Done, nil
}
