package main

import (
	"net/http"
	"strings"

	"github.com/spf13/pflag"
)

var (
	flagsList        *pflag.FlagSet
	flagListExtended bool
	flagListJSON     bool
	flagListColWidth int
)

func init() {
	flagsList = pflag.NewFlagSet("list", pflag.ContinueOnError)
	flagsList.BoolVarP(&flagListExtended, "extended", "x", false, "print records vertically")
	flagsList.Lookup("extended").NoOptDefVal = "true"
	flagsList.BoolVarP(&flagListJSON, "json", "j", false, "print records as JSON")
	flagsList.Lookup("json").NoOptDefVal = "true"
	flagsList.IntVarP(&flagListColWidth, "col-width", "w", 0, "maximum column width; 0 for unlimited")
	RegisterCommand(&Command{
		Name:  "list",
		Help:  "List objects of a particular type, optionally matching a substring.",
		Flags: flagsList,
		Func:  cmdList,
	})
}

var (
	ErrUnknownObjectType = ObserveError{Msg: "unknown object type"}
	ErrListUsage         = ObserveError{Msg: "usage: observe list <object type> [<substring>]"}
	ErrCannotList        = ObserveError{Msg: "cannot list this object type"}
)

func cmdList(cfg *Config, op Output, args []string, hc *http.Client) error {
	if len(args) != 2 && len(args) != 3 {
		return ErrListUsage
	}
	otyp := GetObjectType(args[1])
	if otyp == nil {
		return ErrUnknownObjectType
	}
	if !otyp.CanList() {
		return ErrCannotList
	}
	match := ""
	if len(args) > 2 {
		match = strings.ToLower(args[2])
		op.Debug("match=%s\n", match)
	}
	infos, err := otyp.List(cfg, op, hc)
	if err != nil {
		return NewObserveError(err, "list objects")
	}
	var out TableFormatter
	if flagListJSON {
		out = &JSONFormatter{
			Output:         op,
			ExtendedFormat: flagListExtended,
		}
	} else {
		out = &ColumnFormatter{
			Output:          op,
			OmitLineDrawing: true,
			ColWidth:        flagListColWidth,
			ExtendedFormat:  flagListExtended,
		}
	}
	out.SetColumnNames(otyp.GetPresentationLabels())
	for _, i := range infos {
		if match != "" {
			found := false
			if strings.Contains(strings.ToLower(i.Id), match) {
				found = true
			} else if strings.Contains(strings.ToLower(i.Name), match) {
				found = true
			} else {
			MatchLoop:
				for _, p := range i.Presentation {
					if strings.Contains(strings.ToLower(p), match) {
						found = true
						break MatchLoop
					}
				}
			}
			if !found {
				continue
			}
		}
		out.AddRow(i.Presentation)
	}
	out.Close()
	return nil
}
