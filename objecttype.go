package main

import (
	"fmt"
	"sort"
)

// TODO: merge TypeName/GetProperties() with PropertyType for objects.
type ObjectType interface {
	TypeName() string
	Help() string

	CanList() bool
	List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error)
	CanGet() bool
	Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error)
	CanCreate() bool
	Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error)
	CanUpdate() bool
	Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error)
	CanDelete() bool
	Delete(cfg *Config, op Output, hc httpClient, id string) error

	GetPresentationLabels() []string
	GetProperties() []PropertyDesc
}

type ObjectInfo struct {
	Id           string
	Name         string
	Presentation []string
	Object       ObjectInstance
}

type ObjectInstance interface {
	GetInfo() *ObjectInfo
	GetValues() []PropertyInstance
	PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error
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

// given an input object in "deep" form, unpack it given the propmap, to "flat"
// form.
func propmapObject(src any, propmap PropertyMap) (ret object, err error) {
	ret = make(object)
	for dstKey, srcPath := range propmap {
		ret[dstKey], err = unpackProppath(src, srcPath)
		if err != nil {
			return nil, NewObserveError(err, "property %q", dstKey)
		}
	}
	return ret, nil
}

// given a recursive object, unpack it based on the path, returning the
// terminal "leaf" value.
func unpackProppath(cur any, srcPath proppath) (any, error) {
	for _, key := range srcPath {
		if cur == nil {
			// if we try to index a nil, that's no bueno
			return nil, NewObserveError(ErrNotAnObject, "path %q", srcPath)
		}
		obj, is := cur.(object)
		if !is {
			// if we try to index a non-object, that's no bueno
			return nil, NewObserveError(ErrNotAnObject, "path %q", srcPath)
		}
		cur, is = obj[key]
		if !is {
			// if we don't have the property at all, that's no bueno
			return nil, NewObserveError(ErrNotAnObject, "path %q", srcPath)
		}
		// but it's OK if we have the property, with a nil value, if it's the final leaf
	}
	return cur, nil
}
