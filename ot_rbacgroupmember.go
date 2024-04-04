package main

func init() {
	RegisterObjectType(ObjectTypeRbacgroupmember, &objectRbacgroupmember{})
}

type objectRbacgroupmember struct {
	Id            string
	Description   string
	GroupId       string
	MemberGroupId *string
	MemberUserId  *int64
}

var _ ObjectInstance = &objectRbacgroupmember{}

func (o *objectRbacgroupmember) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           o.Id,
		Name:         o.GroupId,
		Presentation: []string{o.Id, o.GroupId, maybeStr(o.MemberGroupId), maybeIntStr(o.MemberUserId)},
		Object:       o,
	}
}

func (o *objectRbacgroupmember) GetValues() []PropertyInstance {
	props := ObjectTypeRbacgroupmember.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectRbacgroupmember) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeRbacgroupmember struct{}

var ObjectTypeRbacgroupmember ObjectType = &objectTypeRbacgroupmember{}

var propertyDescRbacgroupmember = []PropertyDesc{
	{"id", PropertyTypeString, false, true, func(o any) any { return o.(*objectRbacgroupmember).Id }, func(o any, v any) { o.(*objectRbacgroupmember).Id = v.(string) }},
	{"description", PropertyTypeString, false, false, func(o any) any { return o.(*objectRbacgroupmember).Description }, func(o any, v any) { o.(*objectRbacgroupmember).Description = v.(string) }},
	{"groupid", PropertyTypeString, false, false, func(o any) any { return o.(*objectRbacgroupmember).GroupId }, func(o any, v any) { o.(*objectRbacgroupmember).GroupId = v.(string) }},
	{"membergroupid", PropertyTypeString, false, false, func(o any) any { return maybeStringAny(o.(*objectRbacgroupmember).MemberGroupId) }, func(o any, v any) { o.(*objectRbacgroupmember).MemberGroupId = maybeAnyString(v) }},
	{"memberuserid", PropertyTypeInteger, false, false, func(o any) any { return maybeIntAny(o.(*objectRbacgroupmember).MemberUserId) }, func(o any, v any) { o.(*objectRbacgroupmember).MemberUserId = maybeAnyInt(v) }},
}

func (*objectTypeRbacgroupmember) TypeName() string { return "rbacgroupmember" }
func (*objectTypeRbacgroupmember) Help() string {
	return "A member (user or group) of a group."
}
func (*objectTypeRbacgroupmember) CanList() bool   { return true }
func (*objectTypeRbacgroupmember) CanGet() bool    { return true }
func (*objectTypeRbacgroupmember) CanCreate() bool { return false }
func (*objectTypeRbacgroupmember) CanUpdate() bool { return false }
func (*objectTypeRbacgroupmember) CanDelete() bool { return false }
func (*objectTypeRbacgroupmember) GetPresentationLabels() []string {
	return []string{"id", "groupid", "membergroupid", "memberuserid"}
}
func (*objectTypeRbacgroupmember) GetProperties() []PropertyDesc { return propertyDescRbacgroupmember }

var gqlListRbacgroupmember = compileGqlQuery(`query Rbacgroupmember_List { rbacGroupmembers { id description groupid:groupId membergroupid:memberGroupId memberuserid:memberUserId} }`, "data", "rbacGroupmembers")

func (ot *objectTypeRbacgroupmember) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	obj, err := gqlListRbacgroupmember.query(cfg, op, hc, object{})
	if err != nil || obj == nil {
		return nil, err
	}
	cu := obj.(array)
	var ret []*ObjectInfo
	for _, wks := range cu {
		o := unpackObject(wks.(object), &objectRbacgroupmember{}, ot.TypeName())
		ret = append(ret, o.GetInfo())
	}
	return ret, nil
}

var gqlGetRbacgroupmember = compileGqlQuery(`query Rbacgroupmember_Get_Id($id: ORN!) { rbacGroupmember(id: $id) { id description groupId memberGroupId memberUserId } }`, "data", "rbacGroupmember")

func (ot *objectTypeRbacgroupmember) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetRbacgroupmember.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	return unpackObject(obj.(object), &objectRbacgroupmember{}, ot.TypeName()), nil
}

func (ot *objectTypeRbacgroupmember) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeRbacgroupmember) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeRbacgroupmember) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	return nil
}
