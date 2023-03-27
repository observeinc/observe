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

func (*propertyTypeString) ToString(i any) (string, error) {
	if i == nil {
		return "null", nil
	}
	str, is := i.(string)
	if !is {
		return "", ErrIsNotString
	}
	return strconv.Quote(str), nil
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
