package main

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
)

// TODO: merge TypeName/GetProperties() with PropertyType for objects.
type ObjectType interface {
	TypeName() string
	Help() string

	CanList() bool
	List(cfg *Config, op Output, hc *http.Client) ([]*ObjectInfo, error)
	CanGet() bool
	Get(cfg *Config, op Output, hc *http.Client, id string) (ObjectInstance, error)
	// TODO: Add Create, Put, Delete

	GetPresentationLabels() []string
	GetProperties() []PropertyDesc
}

type ObjectInfo struct {
	Id           string
	Name         string
	Presentation []string
}

type ObjectInstance interface {
	GetInfo() *ObjectInfo
	GetValues() []PropertyInstance
}

var objectTypes = map[string]ObjectType{}

type object = map[string]any
type array = []any

func RegisterObjectType(t ObjectType, tmpl any) {
	if _, has := objectTypes[t.TypeName()]; has {
		panic(fmt.Errorf("Attempt to re-register object type %q", t.TypeName()))
	}
	objectTypes[t.TypeName()] = t
}

func GetObjectType(name string) ObjectType {
	if typ, has := objectTypes[name]; has {
		return typ
	}
	return nil
}

func GetObjectTypes() []ObjectType {
	l := make([]ObjectType, 0, len(objectTypes))
	for _, typ := range objectTypes {
		l = append(l, typ)
	}
	sort.Slice(l, func(i, j int) bool {
		return l[i].TypeName() < l[j].TypeName()
	})
	return l
}

func ForeachObjectType(fn func(ot ObjectType)) {
	for _, typ := range GetObjectTypes() {
		fn(typ)
	}
}

func unpackInfo(obj any, idp PropertyDesc, namep PropertyDesc, props ...PropertyDesc) *ObjectInfo {
	o := obj.(object)
	ret := ObjectInfo{
		Id:   must(idp.Type.Present(idp.Type.FromGQL(o[idp.Name]))),
		Name: must(namep.Type.Present(namep.Type.FromGQL(o[namep.Name]))),
	}
	for _, p := range props {
		ret.Presentation = append(ret.Presentation, must(p.Type.Present(p.Type.FromGQL(o[p.Name]))))
	}
	return &ret
}

// rsp is a map[string]any from a web request
// tgt is the destination to assign to
func unpackObject(rsp object, tgt any, tname string) ObjectInstance {
	// Get the info for this type
	otype := objectTypes[tname]

	for k, v := range rsp {
		pd := getpropdesc(otype, k)
		pd.Setter(tgt, pd.Type.FromGQL(v))
	}
	// Return the result
	return tgt.(ObjectInstance)
}

var fieldTypeInteger = &propertyTypeInteger{}
var fieldTypeString = &propertyTypeString{}
var fieldTypeBoolean = &propertyTypeBoolean{}

var integerType = reflect.TypeOf(int64(0))
var stringType = reflect.TypeOf("")
var booleanType = reflect.TypeOf(false)

func getFieldType(t reflect.Type) PropertyType {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch {
	case t == integerType:
		return fieldTypeInteger
	case t == stringType:
		return fieldTypeString
	case t == booleanType:
		return fieldTypeBoolean
	default:
		panic(fmt.Sprintf("unsupported field type %s", t.Name()))
	}
}

func calcHasConfig(props []PropertyDesc) bool {
	for _, p := range props {
		if !p.IsId && !p.IsComputed {
			return true
		}
	}
	return false
}

func calcHasState(props []PropertyDesc) bool {
	for _, p := range props {
		if !p.IsId && p.IsComputed {
			return true
		}
	}
	return false
}

func writeObjectTypeDocs(op Output, ot ObjectType) {
	props := ot.GetProperties()
	hasConfig := calcHasConfig(props)
	hasState := calcHasState(props)
	fmt.Fprintf(op, "\n%s:\n", ot.TypeName())
	for _, p := range props {
		if p.IsId {
			fmt.Fprintf(op, "  %s: %s\n", p.Name, p.Type.TypeName())
		}
	}
	if hasConfig {
		fmt.Fprint(op, "  config:\n")
		for _, p := range props {
			if !p.IsId && !p.IsComputed {
				fmt.Fprintf(op, "    %s: %s\n", p.Name, p.Type.TypeName())
			}
		}
	}
	if hasState {
		fmt.Fprint(op, "  state:\n")
		for _, p := range props {
			if !p.IsId && p.IsComputed {
				fmt.Fprintf(op, "    %s: %s\n", p.Name, p.Type.TypeName())
			}
		}
	}
}
