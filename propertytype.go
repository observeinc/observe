package main

import (
	"fmt"
	"sort"
)

// TODO: support arrays and objects as properties
type PropertyDesc struct {
	Name       string
	Type       PropertyType
	IsComputed bool
	IsId       bool
	Getter     func(any) any
	Setter     func(any, any)
}

type PropertyInstance interface {
	GetDesc() PropertyDesc
	GetValue() any
}

type propertyInstance struct {
	desc PropertyDesc
	obj  any
}

func (p *propertyInstance) GetDesc() PropertyDesc {
	return p.desc
}

func (p *propertyInstance) GetValue() any {
	return p.desc.Getter(p.obj)
}

type PropertyType interface {
	TypeName() string
	// Present() formats for nice output, but not necessarily machine reading
	Present(any) (string, error)
	// ToString() formats to an unambiguous format that can be parsed with FromString()
	ToString(any) (string, error)
	FromString(string) (any, error)
	FromGQL(any) any
}

var propertyTypes = map[string]PropertyType{}

func RegisterPropertyType(pt PropertyType) {
	if _, has := propertyTypes[pt.TypeName()]; has {
		panic(fmt.Sprintf("attempt to re-register property type %s", pt.TypeName()))
	}
	propertyTypes[pt.TypeName()] = pt
}

func GetPropertyType(ptn string) PropertyType {
	if pt, has := propertyTypes[ptn]; has {
		return pt
	}
	return nil
}

func GetPropertyTypes() []PropertyType {
	l := make([]PropertyType, 0, len(propertyTypes))
	for _, pt := range propertyTypes {
		l = append(l, pt)
	}
	sort.Slice(l, func(i, j int) bool {
		return l[i].TypeName() < l[j].TypeName()
	})
	return l
}

func getpropdesc(ot ObjectType, n string) PropertyDesc {
	props := ot.GetProperties()
	for _, p := range props {
		if p.Name == n {
			return p
		}
	}
	panic(fmt.Errorf("type %q doesn't have property %q", ot.TypeName(), n))
}
