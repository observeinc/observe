package main

func init() {
	RegisterObjectType(ObjectTypeRbacgroup, &objectRbacgroup{})
}

type objectRbacgroup struct {
	Id          string
	Name        string
	Description string
}

var _ ObjectInstance = &objectRbacgroup{}

func (o *objectRbacgroup) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           o.Id,
		Name:         o.Name,
		Presentation: []string{o.Id, o.Name},
		Object:       o,
	}
}

func (o *objectRbacgroup) GetValues() []PropertyInstance {
	props := ObjectTypeRbacgroup.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectRbacgroup) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeRbacgroup struct{}

var ObjectTypeRbacgroup ObjectType = &objectTypeRbacgroup{}

var propertyDescRbacgroup = []PropertyDesc{
	{"id", PropertyTypeString, false, true, func(o any) any { return o.(*objectRbacgroup).Id }, func(o any, v any) { o.(*objectRbacgroup).Id = v.(string) }},
	{"name", PropertyTypeString, false, false, func(o any) any { return o.(*objectRbacgroup).Name }, func(o any, v any) { o.(*objectRbacgroup).Name = v.(string) }},
	{"description", PropertyTypeString, false, false, func(o any) any { return o.(*objectRbacgroup).Description }, func(o any, v any) { o.(*objectRbacgroup).Description = v.(string) }},
}

func (*objectTypeRbacgroup) TypeName() string { return "rbacgroup" }
func (*objectTypeRbacgroup) Help() string {
	return "A group of users and/or other groups for purposes of Role-Based Access Control (RBAC)."
}
func (*objectTypeRbacgroup) CanList() bool                   { return true }
func (*objectTypeRbacgroup) CanGet() bool                    { return true }
func (*objectTypeRbacgroup) CanCreate() bool                 { return false }
func (*objectTypeRbacgroup) CanUpdate() bool                 { return false }
func (*objectTypeRbacgroup) CanDelete() bool                 { return false }
func (*objectTypeRbacgroup) GetPresentationLabels() []string { return []string{"id", "name"} }
func (*objectTypeRbacgroup) GetProperties() []PropertyDesc   { return propertyDescRbacgroup }

var gqlListRbacgroup = compileGqlQuery(`query Rbacgroup_List { rbacGroups { id name description } }`, "data", "rbacGroups")

func (ot *objectTypeRbacgroup) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	obj, err := gqlListRbacgroup.query(cfg, op, hc, object{})
	if err != nil || obj == nil {
		return nil, err
	}
	cu := obj.(array)
	var ret []*ObjectInfo
	for _, wks := range cu {
		o := unpackObject(wks.(object), &objectRbacgroup{}, ot.TypeName())
		ret = append(ret, o.GetInfo())
	}
	return ret, nil
}

var gqlGetRbacgroup = compileGqlQuery(`query Rbacgroup_Get_Id($id: ORN!) { rbacGroup(id: $id) { id name description } }`, "data", "rbacGroup")

func (ot *objectTypeRbacgroup) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetRbacgroup.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	return unpackObject(obj.(object), &objectRbacgroup{}, ot.TypeName()), nil
}

func (ot *objectTypeRbacgroup) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeRbacgroup) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeRbacgroup) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	return nil
}
