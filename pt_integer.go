package main

import "strconv"

func init() {
	RegisterPropertyType(&propertyTypeInteger{})
}

type propertyTypeInteger struct{}

var PropertyTypeInteger PropertyType = &propertyTypeInteger{}

var ErrIsNotInteger = ObserveError{Msg: "value is not an integer"}

func (*propertyTypeInteger) TypeName() string { return "int" }

func (p *propertyTypeInteger) Present(i any) (string, error) {
	return p.ToString(i)
}

func (*propertyTypeInteger) ToString(i any) (string, error) {
	if i == nil {
		return "null", nil
	}
	switch v := i.(type) {
	case *int64:
		if v == nil {
			return "null", nil
		}
		return strconv.FormatInt(*v, 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	default:
		return "", ErrIsNotInteger
	}
}

func (*propertyTypeInteger) FromString(s string) (any, error) {
	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, ErrIsNotInteger
	}
	if s[0] == '+' || s[0] == '0' {
		return nil, ErrIsNotInteger
	}
	return i64, nil
}

func (*propertyTypeInteger) FromGQL(v any) any {
	if v == nil {
		return nil
	}
	return must(strconv.ParseInt(v.(string), 10, 64))
}
