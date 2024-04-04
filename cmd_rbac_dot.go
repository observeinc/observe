package main

import (
	"fmt"

	"github.com/spf13/pflag"
)

var (
	flagsRbacDot    *pflag.FlagSet
	flagRbacDotUser string
	flagRbacDotAll  bool
)

func init() {
	flagsRbacDot = pflag.NewFlagSet("rbac-dot", pflag.ContinueOnError)
	flagsRbacDot.StringVar(&flagRbacDotUser, "user", "", "a user to plot membership graph for")
	flagsRbacDot.BoolVar(&flagRbacDotAll, "all", false, "plot all users/group/statements")
	RegisterCommand(&Command{
		Name:            "rbac-dot",
		Help:            "Generate GraphViz visualizations of RBAC relationships.",
		Func:            cmdRbacDot,
		Flags:           flagsRbacDot,
		Unauthenticated: false,
		Unlisted:        false,
	})
}

var ErrMultiplePlots = ObserveError{Msg: "you can only specify one kind of plot (--user, etc)"}
var ErrMustSpecifyPlot = ObserveError{Msg: "you must specify exactly one plot (--user, etc)"}

func cmdRbacDot(fa FuncArgs) error {
	var todo func(FuncArgs) error
	var prev string
	for _, opt := range []struct {
		oarg  string
		oname string
		fun   func(FuncArgs) error
	}{
		{flagRbacDotUser, "user", cmdRbacDotUser},
	} {
		if opt.oarg != "" {
			if prev != "" {
				return NewObserveError(ErrMultiplePlots, "%s and %s", prev, opt.oname)
			}
			prev = opt.oname
			todo = opt.fun
		}
	}
	if todo == nil {
		if flagRbacDotAll {
			return cmdRbacDotAll(fa)
		}
		return ErrMustSpecifyPlot
	}
	if flagRbacDotAll {
		return NewObserveError(ErrMultiplePlots, "%s and all", prev)
	}
	return todo(fa)
}

func cmdRbacDotUser(fa FuncArgs) error {
	u, err := ObjectTypeUser.Get(fa.cfg, fa.op, fa.hc, flagRbacDotUser)
	if err != nil {
		return err
	}
	user := u.(*objectUser)
	gs, err := ObjectTypeRbacgroup.List(fa.cfg, fa.op, fa.hc)
	if err != nil {
		return err
	}
	groupmap := map[string]*objectRbacgroup{}
	for _, g := range gs {
		gg := g.Object.(*objectRbacgroup)
		groupmap[gg.Id] = gg
	}
	ms, err := ObjectTypeRbacgroupmember.List(fa.cfg, fa.op, fa.hc)
	if err != nil {
		return err
	}
	membermap := map[string]*objectRbacgroupmember{}
	for _, m := range ms {
		mm := m.Object.(*objectRbacgroupmember)
		membermap[mm.Id] = mm
	}
	plotUserGroupsDot(fa.op, user, groupmap, membermap)
	return nil
}

type rbacInstanceState struct {
	users        map[int64]*objectUser
	groups       map[string]*objectRbacgroup
	groupmembers map[string]*objectRbacgroupmember
	statements   map[string]*objectRbacstatement
}

func (ri *rbacInstanceState) fillUsers(fa FuncArgs) error {
	us, err := ObjectTypeUser.List(fa.cfg, fa.op, fa.hc)
	if err != nil {
		return err
	}
	ri.users = map[int64]*objectUser{}
	for _, u := range us {
		uu := u.Object.(*objectUser)
		ri.users[uu.Id] = uu
	}
	return nil
}

func (ri *rbacInstanceState) fillGroups(fa FuncArgs) error {
	gs, err := ObjectTypeRbacgroup.List(fa.cfg, fa.op, fa.hc)
	if err != nil {
		return err
	}
	ri.groups = map[string]*objectRbacgroup{}
	for _, g := range gs {
		gg := g.Object.(*objectRbacgroup)
		ri.groups[gg.Id] = gg
	}
	return nil
}

func (ri *rbacInstanceState) fillGroupMembers(fa FuncArgs) error {
	ms, err := ObjectTypeRbacgroupmember.List(fa.cfg, fa.op, fa.hc)
	if err != nil {
		return err
	}
	ri.groupmembers = map[string]*objectRbacgroupmember{}
	for _, m := range ms {
		mm := m.Object.(*objectRbacgroupmember)
		ri.groupmembers[mm.Id] = mm
	}
	return nil
}

func (ri *rbacInstanceState) fillStatements(fa FuncArgs) error {
	ss, err := ObjectTypeRbacstatement.List(fa.cfg, fa.op, fa.hc)
	if err != nil {
		return err
	}
	ri.statements = map[string]*objectRbacstatement{}
	for _, s := range ss {
		ss := s.Object.(*objectRbacstatement)
		ri.statements[ss.Id] = ss
	}
	return nil
}

func (ri *rbacInstanceState) fillAll(fa FuncArgs) error {
	if err := ri.fillUsers(fa); err != nil {
		return err
	}
	if err := ri.fillGroups(fa); err != nil {
		return err
	}
	if err := ri.fillGroupMembers(fa); err != nil {
		return err
	}
	if err := ri.fillStatements(fa); err != nil {
		return err
	}
	return nil
}

func cmdRbacDotAll(fa FuncArgs) error {
	ri := &rbacInstanceState{}
	if err := ri.fillAll(fa); err != nil {
		return err
	}
	plotFullConnectivityDot(fa.op, ri)
	return nil
}

func plotUserGroupsDot(op Output, user *objectUser, groupmap map[string]*objectRbacgroup, membermap map[string]*objectRbacgroupmember) {
	fmt.Fprintf(op, "digraph {\n")
	fmt.Fprintf(op, "  node [shape=box];\n")
	fmt.Fprintf(op, "  rankdir=LR;\n")
	fmt.Fprintf(op, "  ranksep=1.5;\n")
	fmt.Fprintf(op, "  \"%d\" [label=%q];\n", user.Id, user.Name)

	// seed with user
	groupsToGo := map[string]struct{}{}
	for _, m := range membermap {
		if m.MemberUserId != nil && *m.MemberUserId == user.Id {
			g := groupmap[m.GroupId]
			fmt.Fprintf(op, "  \"%d\" -> %q;\n", user.Id, g.Id)
			groupsToGo[g.Id] = struct{}{}
		}
	}
	// plot transitive memberships
	groupsDone := map[string]struct{}{}
	recursivePlotGroups(op, groupsToGo, groupsDone, groupmap, membermap)
	fmt.Fprintf(op, "}\n")
}

func recursivePlotGroups(op Output, groupsToGo map[string]struct{}, groupsDone map[string]struct{}, groupmap map[string]*objectRbacgroup, membermap map[string]*objectRbacgroupmember) {
	newgg := map[string]struct{}{}
	for g := range groupsToGo {
		if _, has := groupsDone[g]; has {
			continue
		}
		groupsDone[g] = struct{}{}
		gobj := groupmap[g]
		fmt.Fprintf(op, "  %q [label=%q];\n", g, gobj.Name)
		for _, m := range membermap {
			if m.MemberGroupId != nil && *m.MemberGroupId == gobj.Id {
				gg := groupmap[m.GroupId]
				fmt.Fprintf(op, "  %q -> %q;\n", g, gg.Id)
				newgg[gg.Id] = struct{}{}
			}
		}
		recursivePlotGroups(op, newgg, groupsDone, groupmap, membermap)
	}
}

func dotStmtName(s *objectRbacstatement) string {
	if s.ObjectObjectId != nil {
		return fmt.Sprintf("%s obj %d", s.Role, *s.ObjectObjectId)
	}
	if s.ObjectFolderId != nil {
		return fmt.Sprintf("%s fld %d", s.Role, *s.ObjectFolderId)
	}
	if s.ObjectWorkspaceId != nil {
		return fmt.Sprintf("%s wks %d", s.Role, *s.ObjectWorkspaceId)
	}
	if s.ObjectType != nil {
		return fmt.Sprintf("%s %s", s.Role, *s.ObjectType)
	}
	if s.ObjectOwner != nil && *s.ObjectOwner {
		return fmt.Sprintf("%s Owner", s.Role)
	}
	if s.ObjectAll != nil && *s.ObjectAll {
		return fmt.Sprintf("%s All", s.Role)
	}
	return s.Role + " ?"
}

func plotFullConnectivityDot(op Output, ri *rbacInstanceState) {
	fmt.Fprintf(op, "digraph {\n")
	fmt.Fprintf(op, "  newrank=true;\n")
	fmt.Fprintf(op, "  rankdir=LR;\n")
	fmt.Fprintf(op, "  ranksep=10;\n")
	fmt.Fprintf(op, "  subgraph cluster_users {\n")
	fmt.Fprintf(op, "    label=\"Users\";\n")
	fmt.Fprintf(op, "    color=blue;\n")
	fmt.Fprintf(op, "    node [shape=box fixedsize=true width=3 height=1];\n")
	for uid, uobj := range ri.users {
		fmt.Fprintf(op, "    u_%d [label=%q];\n", uid, uobj.Name)
	}
	fmt.Fprintf(op, "  }\n")
	fmt.Fprintf(op, "  subgraph cluster_groups {\n")
	fmt.Fprintf(op, "    label=\"Groups\";\n")
	fmt.Fprintf(op, "    color=green;\n")
	fmt.Fprintf(op, "    node [shape=house fixedsize=true width=3 height=2];\n")
	for gid, gobj := range ri.groups {
		fmt.Fprintf(op, "    %q [label=%q];\n", gid, gobj.Name)
	}
	fmt.Fprintf(op, "  }\n")
	fmt.Fprintf(op, "  All [shape=doublecircle width=2 height=2 fixedsize=true];\n")
	fmt.Fprintf(op, "  subgraph cluster_statements {\n")
	fmt.Fprintf(op, "    label=\"Statements\";\n")
	fmt.Fprintf(op, "    color=red;\n")
	fmt.Fprintf(op, "    node [shape=oval fixedsize=true width=2 height=1];\n")
	for sid, sobj := range ri.statements {
		fmt.Fprintf(op, "    %q [label=%q];\n", sid, dotStmtName(sobj))
	}
	fmt.Fprintf(op, "  }\n")
	for _, gmobj := range ri.groupmembers {
		if gmobj.MemberUserId != nil && ri.users[*gmobj.MemberUserId] != nil {
			fmt.Fprintf(op, "  u_%d -> %q [weight=1];\n", *gmobj.MemberUserId, gmobj.GroupId)
		}
	}
	for _, gmobj := range ri.groupmembers {
		if gmobj.MemberGroupId != nil {
			fmt.Fprintf(op, "  %q -> %q [weight=2];\n", *gmobj.MemberGroupId, gmobj.GroupId)
		}
	}
	for _, sobj := range ri.statements {
		if sobj.SubjectUser != nil {
			fmt.Fprintf(op, "  %q -> u_%d [weight=3];\n", sobj.Id, *sobj.SubjectUser)
		} else if sobj.SubjectGroup != nil {
			fmt.Fprintf(op, "  %q -> %q [weight=2];\n", sobj.Id, *sobj.SubjectGroup)
		} else {
			fmt.Fprintf(op, "  %q -> All [weight=1];\n", sobj.Id)
		}
	}
	fmt.Fprintf(op, "}\n")
}
