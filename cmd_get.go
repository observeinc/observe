package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/pflag"
)

var (
	flagsGet *pflag.FlagSet
)

func init() {
	flagsGet = pflag.NewFlagSet("get", pflag.ContinueOnError)
	RegisterCommand(&Command{
		Name:  "get",
		Help:  "Get the state of a particular object identified by type and id. Use `observe help --objects` for a list of object types.",
		Flags: flagsGet,
		Func:  cmdGet,
	})
}

var (
	ErrGetUsage  = ObserveError{Msg: "usage: observe get <object type> <object id>"}
	ErrCannotGet = ObserveError{Msg: "cannot get this object type"}
)

func cmdGet(cfg *Config, op Output, args []string, hc *http.Client) error {
	if len(args) != 3 {
		return ErrGetUsage
	}
	otyp := GetObjectType(args[1])
	if otyp == nil {
		return ErrUnknownObjectType
	}
	if !otyp.CanGet() {
		return ErrCannotGet
	}
	obj, err := otyp.Get(cfg, op, hc, args[2])
	if err != nil {
		return NewObserveError(err, "get objects")
	}
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
