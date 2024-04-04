package main

import (
	"strconv"
)

func init() {
	RegisterPropertyType(&propertyTypeBoolean{})
}

var ErrIsNotBoolean = ObserveError{Msg: "value is not a boolean"}

type propertyTypeBoolean struct{}

var PropertyTypeBoolean PropertyType = &propertyTypeBoolean{}

func (*propertyTypeBoolean) TypeName() string { return "boolean" }

func (p *propertyTypeBoolean) Present(i any) (string, error) {
	return p.ToString(i)
}

func (*propertyTypeBoolean) ToString(i any) (string, error) {
	if i == nil {
		return "", nil
	}
	switch v := i.(type) {
	case *bool:
		if v == nil {
			return "", nil
		}
		return strconv.FormatBool(*v), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", ErrIsNotBoolean
	}
}

func (*propertyTypeBoolean) FromString(s string) (any, error) {
	if s == "" {
		return nil, nil
	}
	if s == "true" {
		return true, nil
	}
	if s == "false" {
		return false, nil
	}
	return nil, ErrIsNotBoolean
}

func (*propertyTypeBoolean) FromGQL(v any) any {
	if v == nil {
		return nil
	}
	return v.(bool)
}
