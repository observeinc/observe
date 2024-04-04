package main

import (
	"strconv"
)

func init() {
	RegisterObjectType(ObjectTypeUser, &objectUser{})
}

type objectUser struct {
	Id     int64
	Name   string
	Email  string
	Status string
	Role   string
}

var _ ObjectInstance = &objectUser{}

func (o *objectUser) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           strconv.FormatInt(o.Id, 10),
		Name:         o.Name,
		Presentation: []string{strconv.FormatInt(o.Id, 10), o.Name, o.Email},
		Object:       o,
	}
}

func (o *objectUser) GetValues() []PropertyInstance {
	props := ObjectTypeUser.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectUser) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeUser struct{}

var ObjectTypeUser ObjectType = &objectTypeUser{}

var propertyDescUser = []PropertyDesc{
	{"id", PropertyTypeInteger, false, true, func(o any) any { return o.(*objectUser).Id }, func(o any, v any) { o.(*objectUser).Id = v.(int64) }},
	{"name", PropertyTypeString, false, false, func(o any) any { return o.(*objectUser).Name }, func(o any, v any) { o.(*objectUser).Name = v.(string) }},
	{"email", PropertyTypeString, false, false, func(o any) any { return o.(*objectUser).Email }, func(o any, v any) { o.(*objectUser).Email = v.(string) }},
	{"status", PropertyTypeString, false, false, func(o any) any { return o.(*objectUser).Status }, func(o any, v any) { o.(*objectUser).Status = v.(string) }},
	{"role", PropertyTypeString, false, false, func(o any) any { return o.(*objectUser).Role }, func(o any, v any) { o.(*objectUser).Role = v.(string) }},
}

func (*objectTypeUser) TypeName() string { return "user" }
func (*objectTypeUser) Help() string {
	return "A human or serviceaccount user of the Observe tenant."
}
func (*objectTypeUser) CanList() bool                   { return true }
func (*objectTypeUser) CanGet() bool                    { return true }
func (*objectTypeUser) CanCreate() bool                 { return false }
func (*objectTypeUser) CanUpdate() bool                 { return false }
func (*objectTypeUser) CanDelete() bool                 { return false }
func (*objectTypeUser) GetPresentationLabels() []string { return []string{"id", "name", "email"} }
func (*objectTypeUser) GetProperties() []PropertyDesc   { return propertyDescUser }

var gqlListUser = compileGqlQuery(`query User_List { currentCustomer { users { id name:label email status role } } }`, "data", "currentCustomer", "users")

func (ot *objectTypeUser) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	obj, err := gqlListUser.query(cfg, op, hc, object{})
	if err != nil || obj == nil {
		return nil, err
	}
	cu := obj.(array)
	var ret []*ObjectInfo
	for _, wks := range cu {
		o := unpackObject(wks.(object), &objectUser{}, ot.TypeName())
		ret = append(ret, o.GetInfo())
	}
	return ret, nil
}

var gqlGetUser = compileGqlQuery(`query User_Get_Id($id: UserId!) { user(id: $id) { id name:label email status role } }`, "data", "user")

func (ot *objectTypeUser) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetUser.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	return unpackObject(obj.(object), &objectUser{}, ot.TypeName()), nil
}

func (ot *objectTypeUser) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeUser) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeUser) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	return nil
}
