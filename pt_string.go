package main

import (
	"strconv"
)

func init() {
	RegisterPropertyType(&propertyTypeString{})
}

var ErrIsNotString = ObserveError{Msg: "value is not a string"}

type propertyTypeString struct{}

var PropertyTypeString PropertyType = &propertyTypeString{}

func (*propertyTypeString) TypeName() string { return "string" }

func (*propertyTypeString) Present(i any) (string, error) {
	if i == nil {
		return "", nil
	}
	str, is := i.(string)
	if !is {
		return "", ErrIsNotString
	}
	return str, nil
}

func (*propertyTypeString) ToString(i any) (string, error) {
	if i == nil {
		return "", nil
	}
	switch v := i.(type) {
	case *string:
		if v == nil {
			return "", nil
		}
		return strconv.Quote(*v), nil
	case string:
		return strconv.Quote(v), nil
	default:
		return "", ErrIsNotString
	}
}

func (*propertyTypeString) FromString(s string) (any, error) {
	if str, err := strconv.Unquote(s); err != nil {
		return nil, ErrIsNotString
	} else {
		return str, nil
	}
}

func (*propertyTypeString) FromGQL(v any) any {
	if v == nil {
		return nil
	}
	return v.(string)
}
