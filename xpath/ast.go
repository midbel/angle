package xpath

type Axis string

type Step struct {
	axis Axis
	Expr
}

func (s *Step) Eval(ctx Context) (Sequence, error) {
	return nil, nil
}

type nameTest struct {
	name xml.Name
}

func (t *nameTest) Eval(ctx Context) bool {
	return false
}

type attrTest struct {
	name xml.Name
}

func (t *attrTest) Eval(ctx Context) bool {
	return false
}
