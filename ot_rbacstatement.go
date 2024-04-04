package main

func init() {
	RegisterObjectType(ObjectTypeRbacstatement, &objectRbacstatement{})
}

type objectRbacstatement struct {
	Id                string
	Description       string
	SubjectGroup      *string
	SubjectUser       *int64
	SubjectAll        *bool
	ObjectObjectId    *int64
	ObjectFolderId    *int64
	ObjectWorkspaceId *int64
	ObjectType        *string
	ObjectName        *string
	ObjectOwner       *bool
	ObjectAll         *bool
	Role              string
}

var _ ObjectInstance = &objectRbacstatement{}

func (o *objectRbacstatement) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           o.Id,
		Name:         o.Description,
		Presentation: []string{o.Id, o.Description},
		Object:       o,
	}
}

func (o *objectRbacstatement) GetValues() []PropertyInstance {
	props := ObjectTypeRbacstatement.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectRbacstatement) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objecttypeRbacstatement struct{}

var ObjectTypeRbacstatement ObjectType = &objecttypeRbacstatement{}

var remapRbacstatement = remap{
	"subjectgroupid":    "subject.groupId",
	"subjectuserid":     "subject.userId",
	"subjectAll":        "subject.all",
	"objectobjectid":    "object.objectId",
	"objectfolderid":    "object.folderId",
	"objectworkspaceid": "object.workspaceId",
	"objecttype":        "object.type",
	"objectname":        "object.name",
	"objectowner":       "object.owner",
	"objectall":         "object.all",
}

var propertyDescRbacstatement = []PropertyDesc{
	{"id", PropertyTypeString, false, true, func(o any) any { return o.(*objectRbacstatement).Id }, func(o any, v any) { o.(*objectRbacstatement).Id = v.(string) }},
	{"description", PropertyTypeString, false, false, func(o any) any { return o.(*objectRbacstatement).Description }, func(o any, v any) { o.(*objectRbacstatement).Description = v.(string) }},
	{"subjectgroupid", PropertyTypeString, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).SubjectGroup) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectRbacstatement).SubjectGroup = vp
	}},
	{"subjectuserid", PropertyTypeInteger, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).SubjectUser) }, func(o any, v any) {
		var vp *int64
		if v != nil {
			vs := v.(int64)
			vp = &vs
		}
		o.(*objectRbacstatement).SubjectUser = vp
	}},
	{"subjectAll", PropertyTypeBoolean, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).SubjectAll) }, func(o any, v any) {
		var vp *bool
		if v != nil {
			vs := v.(bool)
			vp = &vs
		}
		o.(*objectRbacstatement).SubjectAll = vp
	}},
	{"objectobjectid", PropertyTypeInteger, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).ObjectObjectId) }, func(o any, v any) {
		var vp *int64
		if v != nil {
			vs := v.(int64)
			vp = &vs
		}
		o.(*objectRbacstatement).ObjectObjectId = vp
	}},
	{"objectfolderid", PropertyTypeInteger, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).ObjectFolderId) }, func(o any, v any) {
		var vp *int64
		if v != nil {
			vs := v.(int64)
			vp = &vs
		}
		o.(*objectRbacstatement).ObjectFolderId = vp
	}},
	{"objectworkspaceid", PropertyTypeInteger, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).ObjectWorkspaceId) }, func(o any, v any) {
		var vp *int64
		if v != nil {
			vs := v.(int64)
			vp = &vs
		}
		o.(*objectRbacstatement).ObjectWorkspaceId = vp
	}},
	{"objecttype", PropertyTypeString, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).ObjectType) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectRbacstatement).ObjectType = vp
	}},
	{"objectname", PropertyTypeString, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).ObjectName) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectRbacstatement).ObjectName = vp
	}},
	{"objectowner", PropertyTypeBoolean, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).ObjectOwner) }, func(o any, v any) {
		var vp *bool
		if v != nil {
			vs := v.(bool)
			vp = &vs
		}
		o.(*objectRbacstatement).ObjectOwner = vp
	}},
	{"objectall", PropertyTypeBoolean, false, false, func(o any) any { return maybe(o.(*objectRbacstatement).ObjectAll) }, func(o any, v any) {
		var vp *bool
		if v != nil {
			vs := v.(bool)
			vp = &vs
		}
		o.(*objectRbacstatement).ObjectAll = vp
	}},
	{"role", PropertyTypeString, false, false, func(o any) any { return o.(*objectRbacstatement).Role }, func(o any, v any) { o.(*objectRbacstatement).Role = v.(string) }},
}

func (*objecttypeRbacstatement) TypeName() string { return "rbacstatement" }
func (*objecttypeRbacstatement) Help() string {
	return "A rule that binds some role (permission) to some object or object type for some user or group."
}
func (*objecttypeRbacstatement) CanList() bool                   { return true }
func (*objecttypeRbacstatement) CanGet() bool                    { return true }
func (*objecttypeRbacstatement) CanCreate() bool                 { return false }
func (*objecttypeRbacstatement) CanUpdate() bool                 { return false }
func (*objecttypeRbacstatement) CanDelete() bool                 { return false }
func (*objecttypeRbacstatement) GetPresentationLabels() []string { return []string{"id", "name"} }
func (*objecttypeRbacstatement) GetProperties() []PropertyDesc   { return propertyDescRbacstatement }

var gqlListRbacstatement = compileGqlQuery(
	`query Rbacstatement_List { rbacStatements { id description subject { userId groupId all } object { objectId folderId workspaceId type name owner all } role } }`, "data", "rbacStatements").
	WithRemap(remapRbacstatement)

func (ot *objecttypeRbacstatement) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	obj, err := gqlListRbacstatement.query(cfg, op, hc, object{})
	if err != nil || obj == nil {
		return nil, err
	}
	cu := obj.(array)
	var ret []*ObjectInfo
	for _, wks := range cu {
		o := unpackObject(wks.(object), &objectRbacstatement{}, ot.TypeName())
		ret = append(ret, o.GetInfo())
	}
	return ret, nil
}

var gqlGetRbacstatement = compileGqlQuery(
	`query Rbacstatement_Get_Id($id: ORN!) { rbacStatement(id: $id) { id description subject { userId groupId all } object { objectId folderId workspaceId type name owner all } role } }`, "data", "rbacStatement").
	WithRemap(remapRbacstatement)

func (ot *objecttypeRbacstatement) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := gqlGetRbacstatement.query(cfg, op, hc, object{"id": id})
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	return unpackObject(obj.(object), &objectRbacstatement{}, ot.TypeName()), nil
}

func (ot *objecttypeRbacstatement) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objecttypeRbacstatement) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objecttypeRbacstatement) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	return nil
}
