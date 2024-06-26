package main

import (
	"regexp"
	"strconv"
)

func init() {
	RegisterPropertyType(&propertyTypeORN{})
}

var reORN = regexp.MustCompile("^o:[^:]*:[^:]*:[^:]*:[^:*](/.*)?$")

var ErrIsNotORN = ObserveError{Msg: "value is not an orn"}

type propertyTypeORN struct{}

func (*propertyTypeORN) TypeName() string { return "orn" }

func (*propertyTypeORN) Present(i any) (string, error) {
	if i == nil {
		return "", nil
	}
	str, is := i.(string)
	if !is {
		return "", ErrIsNotORN
	}
	if !reORN.MatchString(str) {
		return "", ErrIsNotORN
	}
	return str, nil
}

func (*propertyTypeORN) ToString(i any) (string, error) {
	if i == nil {
		return "", nil
	}
	switch v := i.(type) {
	case *string:
		if v == nil {
			return "", nil
		}
		if !reORN.MatchString(*v) {
			return "", ErrIsNotORN
		}
		return strconv.Quote(*v), nil
	case string:
		if !reORN.MatchString(v) {
			return "", ErrIsNotORN
		}
		return strconv.Quote(v), nil
	default:
		return "", ErrIsNotString
	}
}

func (*propertyTypeORN) FromString(s string) (any, error) {
	if s == "" {
		return nil, nil
	}
	if len(s) < 2 {
		return nil, ErrIsNotORN
	}
	if s[0] != '"' || s[len(s)-1] != '"' {
		return nil, ErrIsNotORN
	}
	if str, err := strconv.Unquote(s); err != nil {
		return nil, ErrIsNotORN
	} else if !reORN.MatchString(str) {
		return nil, ErrIsNotORN
	} else {
		return str, nil
	}
}

func (*propertyTypeORN) FromGQL(v any) any {
	if v == nil {
		return nil
	}
	return v.(string)
}
