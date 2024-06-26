package main

import (
	"strconv"
)

func init() {
	RegisterObjectType(ObjectTypeWorkspace, &objectWorkspace{})
}

type objectWorkspace struct {
	workspaceObject
	Timezone string
}

var _ ObjectInstance = &objectWorkspace{}

func (o *objectWorkspace) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           strconv.FormatInt(o.Id, 10),
		Name:         o.Name,
		Presentation: []string{strconv.FormatInt(o.Id, 10), o.Name},
		Object:       o,
	}
}

func (o *objectWorkspace) GetValues() []PropertyInstance {
	props := ObjectTypeWorkspace.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectWorkspace) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeWorkspace struct{}

var ObjectTypeWorkspace ObjectType = &objectTypeWorkspace{}

var propertyDescWorkspace = []PropertyDesc{
	{"id", PropertyTypeInteger, false, true, func(o any) any { return o.(*objectWorkspace).Id }, func(o any, v any) { o.(*objectWorkspace).Id = v.(int64) }},
	{"name", PropertyTypeString, false, false, func(o any) any { return o.(*objectWorkspace).Name }, func(o any, v any) { o.(*objectWorkspace).Name = v.(string) }},
	{"timezone", PropertyTypeString, false, false, func(o any) any { return o.(*objectWorkspace).Timezone }, func(o any, v any) { o.(*objectWorkspace).Timezone = v.(string) }},
}

func (*objectTypeWorkspace) TypeName() string { return "workspace" }
func (*objectTypeWorkspace) Help() string {
	return "A workspace organizes most other objects, such as datasets."
}
func (*objectTypeWorkspace) CanList() bool                   { return true }
func (*objectTypeWorkspace) CanGet() bool                    { return true }
func (*objectTypeWorkspace) CanCreate() bool                 { return false }
func (*objectTypeWorkspace) CanUpdate() bool                 { return false }
func (*objectTypeWorkspace) CanDelete() bool                 { return false }
func (*objectTypeWorkspace) GetPresentationLabels() []string { return []string{"id", "name"} }
func (*objectTypeWorkspace) GetProperties() []PropertyDesc   { return propertyDescWorkspace }

var gqlListWorkspace = compileGqlQuery(`query Workspace_List { currentUser { workspaces { id name:label } } }`, "data", "currentUser", "workspaces")

func (ot *objectTypeWorkspace) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	obj, err := gqlListWorkspace.query(cfg, op, hc, object{})
	if err != nil || obj == nil {
		return nil, err
	}
	cu := obj.(array)
	var ret []*ObjectInfo
	for _, wks := range cu {
		o := unpackObject(wks.(object), &objectWorkspace{}, ot.TypeName())
		ret = append(ret, o.GetInfo())
	}
	return ret, nil
}

var gqlGetWorkspace = compileGqlQuery(`query Workspace_Get_Id($id: ObjectId!) { workspace(id: $id) { id name:label timezone } }`, "data", "workspace")

func (ot *objectTypeWorkspace) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetWorkspace.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	return unpackObject(obj.(object), &objectWorkspace{}, ot.TypeName()), nil
}

func (ot *objectTypeWorkspace) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeWorkspace) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeWorkspace) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	return nil
}
