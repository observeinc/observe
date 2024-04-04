package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	flagsGet *pflag.FlagSet
)

func init() {
	flagsGet = pflag.NewFlagSet("get", pflag.ContinueOnError)
	RegisterCommand(&Command{
		Name:  "get",
		Help:  "Get the state of a particular object identified by type and id.",
		Flags: flagsGet,
		Func:  cmdGet,
	})
}

var (
	ErrGetUsage  = ObserveError{Msg: "usage: observe get <object type> <object id>"}
	ErrCannotGet = ObserveError{Msg: "cannot get this object type"}
)

func cmdGet(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ErrGetUsage
	}
	otyp := GetObjectType(fa.args[1])
	if otyp == nil {
		return ErrUnknownObjectType
	}
	if !otyp.CanGet() {
		return ErrCannotGet
	}

	obj, err := otyp.Get(fa.cfg, fa.op, fa.hc, fa.args[2])
	if err != nil {
		return NewObserveError(err, "get object type:%s", otyp.TypeName())
	}
	return obj.PrintToYaml(fa.op, otyp, obj)
}

// printToYamlFromJson converts `object` to `yaml`, sorting the top level fields in a deterministic way.
func printToYamlFromJson(op Output, otyp ObjectType, obj object) error {

	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)

	// TODO(OB-20759): make a better way to define ordering by preserving the original order returned by the api.
	parts := []string{"type", "id", "config", "state", "meta"}
	for _, key := range parts {
		val, ok := obj[key]
		if !ok {
			continue
		}
		if err := enc.Encode(map[string]any{key: val}); err != nil {
			return err
		}
	}
	str := buf.String()
	strs := strings.Split(str, "\n---\n")
	str = strings.Join(strs, "\n")
	fmt.Fprint(op, str)
	return nil
}

// Note: we probably should use yaml/v3, or some more introspection based
// discovery.
func printToYamlFromObjectInstance(op Output, otyp ObjectType, obj ObjectInstance) error {
	fmt.Fprintf(op, "object:\n")
	fmt.Fprintf(op, "  type: %q\n", otyp.TypeName())
	vals := obj.GetValues()
	props := otyp.GetProperties()
	hasConfig := calcHasConfig(props)
	hasState := calcHasState(props)
	for _, v := range vals {
		p := v.GetDesc()
		if p.IsId {
			vstr, e := p.Type.ToString(v.GetValue())
			if e != nil {
				vstr = strconv.Quote(e.Error())
			}
			fmt.Fprintf(op, "  %s: %s\n", p.Name, vstr)
		}
	}
	if hasConfig {
		fmt.Fprintf(op, "  config:\n")
		for _, v := range vals {
			p := v.GetDesc()
			if !p.IsId && !p.IsComputed {
				vstr, e := p.Type.ToString(v.GetValue())
				if e != nil {
					vstr = strconv.Quote(e.Error())
				}
				fmt.Fprintf(op, "    %s: %s\n", p.Name, vstr)
			}
		}
	}
	if hasState {
		fmt.Fprintf(op, "  state:\n")
		for _, v := range vals {
			p := v.GetDesc()
			if !p.IsId && p.IsComputed {
				vstr, e := p.Type.ToString(v.GetValue())
				if e != nil {
					vstr = strconv.Quote(e.Error())
				}
				fmt.Fprintf(op, "    %s: %s\n", p.Name, vstr)
			}
		}
	}
	return nil
}
