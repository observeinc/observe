package main

import "strconv"

func init() {
	RegisterPropertyType(&propertyTypeInteger{})
}

type propertyTypeInteger struct{}

var PropertyTypeInteger PropertyType = &propertyTypeInteger{}

var ErrIsNotInteger = ObserveError{Msg: "value is not an integer"}

func (*propertyTypeInteger) TypeName() string { return "int" }

func (*propertyTypeInteger) ToString(i any) (string, error) {
	if i == nil {
		return "null", nil
	}
	i64, is := i.(int64)
	if !is {
		return "", ErrIsNotInteger
	}
	return strconv.FormatInt(i64, 10), nil
}

func (*propertyTypeInteger) FromString(s string) (any, error) {
	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
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
