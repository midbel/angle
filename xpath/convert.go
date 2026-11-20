package xpath

import "strconv"

func convertTo(left, right Item) (any, any, error) {
	e, err := atomize(left.Value())
	if err != nil {
		return nil, nil, err
	}
	r, err := atomize(right.Value())
	if err != nil {
		return nil, nil, err
	}
	var x any
	switch e := e.(type) {
	case untyped:
		x, err = convertUntyped(e, r)
	case float64:
		x, err = convertFloat(e, r)
	case int64:
		x, err = convertInt(e, r)
	case bool:
		x, err = convertBool(e, r)
	case string:
		x, err = convertString(e, r)
	default:
		return nil, nil, ErrConvert
	}
	return x, r, err
}

func promoteTo(left, right any) (any, any, error) {
	switch e := left.(type) {
	case int64:
		if _, ok := right.(float64); ok {
			return float64(e), right, nil
		}
	case float64:
		if r, ok := right.(int64); ok {
			return left, float64(r), nil
		}
	default:
		return nil, nil, ErrType
	}
	return left, right, nil
}

func convertUntyped(value untyped, target any) (any, error) {
	switch target.(type) {
	case untyped:
		return value, nil
	case float64, int64:
		v, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			err = ErrConvert
		}
		return v, err
	case string:
		return string(value), nil
	case bool:
		b, err := strconv.ParseBool(string(value))
		if err != nil {
			err = ErrConvert
		}
		return b, err
	default:
		return nil, ErrConvert
	}
}

func convertFloat(value float64, target any) (any, error) {
	switch target.(type) {
	case float64:
		return value, nil
	case string:
		return strconv.FormatFloat(value, 'g', -1, 64), nil
	default:
		return nil, ErrConvert
	}
}

func convertInt(value int64, target any) (any, error) {
	switch target.(type) {
	case int64:
		return value, nil
	case string:
		return strconv.FormatInt(value, 10), nil
	default:
		return nil, ErrConvert
	}
}

func convertBool(value bool, target any) (any, error) {
	switch target.(type) {
	case bool:
		return value, nil
	default:
		return nil, ErrConvert
	}
}

func convertString(value string, target any) (any, error) {
	switch target.(type) {
	case string:
		return value, nil
	default:
		return nil, ErrConvert
	}
}

func isNumeric(value any) bool {
	switch value.(type) {
	case int64, float64:
		return true
	default:
		return false
	}
}
