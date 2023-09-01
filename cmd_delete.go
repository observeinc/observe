package main

import "fmt"

var (
	ErrDeleteUsage  = ObserveError{Msg: "usage: observe delete <object type> <object id>"}
	ErrCannotDelete = ObserveError{Msg: "cannot delete this object type"}
)

func init() {
	RegisterCommand(&Command{
		Name: "delete",
		Help: "Delete an object identified by type and id.",
		Func: cmdDelete,
	})
}

func cmdDelete(fa FuncArgs) error {
	if len(fa.args) < 2 {
		return ErrDeleteUsage
	}
	otyp := GetObjectType(fa.args[1])
	if otyp == nil {
		return ErrUnknownObjectType
	}
	if !otyp.CanDelete() {
		return ErrCannotDelete
	}
	// TODO: Add support for delete --name
	if len(fa.args) != 3 {
		return ErrDeleteUsage
	}
	err := otyp.Delete(fa.cfg, fa.op, fa.hc, fa.args[2])
	if err != nil {
		return NewObserveError(err, "delete %s", otyp.TypeName())
	}
	_, err = fmt.Fprintf(fa.op, "deleted %s\n", fa.args[2])
	return err
}
