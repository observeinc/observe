package main

import (
	"net/http"
	"strings"

	"github.com/spf13/pflag"
)

var (
	flagsList        *pflag.FlagSet
	flagListExtended bool
	flagListColWidth int
)

func init() {
	flagsList = pflag.NewFlagSet("list", pflag.ContinueOnError)
	flagsList.BoolVarP(&flagListExtended, "extended", "x", false, "print records vertically")
	flagsList.Lookup("extended").NoOptDefVal = "true"
	flagsList.IntVarP(&flagListColWidth, "col-width", "w", 0, "maximum column width; 0 for unlimited")
	RegisterCommand(&Command{
		Name:  "list",
		Help:  "List objects of a particular type. Use `observe help --objects` for a list of object types.",
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
	out := &ColumnFormatter{
		Output:          op,
		OmitLineDrawing: true,
		ColWidth:        flagListColWidth,
		ExtendedFormat:  flagListExtended,
	}
	out.SetColumnNames(append([]string{"id", "name"}, otyp.GetPresentationLabels()...))
	for _, i := range infos {
		if match != "" && !(strings.Contains(strings.ToLower(i.Id), match) || strings.Contains(strings.ToLower(i.Name), match)) {
			continue
		}
		out.AddRow(append([]string{i.Id, i.Name}, i.Presentation...))
	}
	out.Close()
	return nil
}
