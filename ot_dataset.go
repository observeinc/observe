package main

import (
	"net/http"
	"strconv"
)

func init() {
	RegisterObjectType(ObjectTypeDataset, &objectDataset{})
}

type objectDataset struct {
	Id             int64
	Name           string
	Workspace      int64
	Path           string
	Kind           string
	Description    *string
	ValidFromField *string
	ValidToField   *string
	LabelField     *string
	IconUrl        *string
	Version        string
	UpdatedDate    string
	PathCost       *int64
	ManagedById    *int64
	FolderId       int64
	// todo: compound property types
	// CompilationError *CompilationError
	// PrimaryKey []string
	// Keys [][]string
	// Typedef    []DatasetColumn
	// ForeignKeys []ForeignKey
	// RelatedKeys []RelatedKey
}

var _ ObjectInstance = &objectDataset{}

func (o *objectDataset) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           strconv.FormatInt(o.Id, 10),
		Name:         o.Name,
		Presentation: []string{strconv.FormatInt(o.Id, 10), o.Path},
	}
}

func (o *objectDataset) GetValues() []PropertyInstance {
	props := ObjectTypeDataset.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

type objectTypeDataset struct{}

var ObjectTypeDataset ObjectType = &objectTypeDataset{}

// TODO: We should extract the input/output structs from GQL schema and use
// guided codegen to generate these object type definitions.
var propertyDescDataset = []PropertyDesc{
	{"id", PropertyTypeInteger, false, true, func(o any) any { return o.(*objectDataset).Id }, func(o any, v any) { o.(*objectDataset).Id = v.(int64) }},
	{"path", PropertyTypeString, true, false, func(o any) any { return o.(*objectDataset).Path }, func(o any, v any) { o.(*objectDataset).Path = v.(string) }},
	{"workspace", PropertyTypeInteger, false, false, func(o any) any { return o.(*objectDataset).Workspace }, func(o any, v any) { o.(*objectDataset).Workspace = v.(int64) }},
	{"folderId", PropertyTypeInteger, false, false, func(o any) any { return o.(*objectDataset).FolderId }, func(o any, v any) { o.(*objectDataset).FolderId = v.(int64) }},
	{"name", PropertyTypeString, false, false, func(o any) any { return o.(*objectDataset).Name }, func(o any, v any) { o.(*objectDataset).Name = v.(string) }},
	{"kind", PropertyTypeString, true, false, func(o any) any { return o.(*objectDataset).Kind }, func(o any, v any) { o.(*objectDataset).Kind = v.(string) }},
	{"description", PropertyTypeString, false, false, func(o any) any { return maybe(o.(*objectDataset).Description) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectDataset).Description = vp
	}},
	{"managedById", PropertyTypeInteger, true, false, func(o any) any { return maybe(o.(*objectDataset).ManagedById) }, func(o any, v any) {
		var vp *int64
		if v != nil {
			vs := v.(int64)
			vp = &vs
		}
		o.(*objectDataset).ManagedById = vp
	}},
	{"pathCost", PropertyTypeInteger, false, false, func(o any) any { return maybe(o.(*objectDataset).PathCost) }, func(o any, v any) {
		var vp *int64
		if v != nil {
			vs := v.(int64)
			vp = &vs
		}
		o.(*objectDataset).PathCost = vp
	}},
	{"validFromField", PropertyTypeString, true, false, func(o any) any { return maybe(o.(*objectDataset).ValidFromField) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectDataset).ValidFromField = vp
	}},
	{"validToField", PropertyTypeString, true, false, func(o any) any { return maybe(o.(*objectDataset).ValidToField) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectDataset).ValidToField = vp
	}},
	{"labelField", PropertyTypeString, true, false, func(o any) any { return maybe(o.(*objectDataset).LabelField) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectDataset).LabelField = vp
	}},
	{"iconUrl", PropertyTypeString, false, false, func(o any) any { return maybe(o.(*objectDataset).IconUrl) }, func(o any, v any) {
		var vp *string
		if v != nil {
			vs := v.(string)
			vp = &vs
		}
		o.(*objectDataset).IconUrl = vp
	}},
	{"version", PropertyTypeString, true, false, func(o any) any { return o.(*objectDataset).Version }, func(o any, v any) { o.(*objectDataset).Version = v.(string) }},
	{"updatedDate", PropertyTypeString, true, false, func(o any) any { return o.(*objectDataset).UpdatedDate }, func(o any, v any) { o.(*objectDataset).UpdatedDate = v.(string) }},
}

func (*objectTypeDataset) TypeName() string { return "dataset" }
func (*objectTypeDataset) Help() string {
	return "A dataset contains processed data ready to be queried."
}
func (*objectTypeDataset) CanList() bool                   { return true }
func (*objectTypeDataset) CanGet() bool                    { return true }
func (*objectTypeDataset) GetPresentationLabels() []string { return []string{"id", "path"} }
func (*objectTypeDataset) GetProperties() []PropertyDesc {
	return propertyDescDataset
}

func (ot *objectTypeDataset) List(cfg *Config, op Output, hc *http.Client) ([]*ObjectInfo, error) {
	obj, err := gqlQuery(cfg, op, hc, `query Dataset_List { datasetSearch { dataset { id name path } } }`, object{}, "data", "datasetSearch")
	if err != nil || obj == nil {
		return nil, err
	}
	cu := obj.(array)
	var ret []*ObjectInfo
	idp := getpropdesc(ot, "id")
	namep := getpropdesc(ot, "name")
	pathp := getpropdesc(ot, "path")
	for _, ds := range cu {
		ret = append(ret, unpackInfo(ds.(object)["dataset"], idp, namep, idp, pathp))
	}
	return ret, nil
}

func (ot *objectTypeDataset) Get(cfg *Config, op Output, hc *http.Client, id string) (ObjectInstance, error) {
	obj, err := gqlQuery(cfg, op, hc, `query Dataset_Get_Id($id: ObjectId!) { dataset(id: $id) { id name:label workspace:workspaceId path kind description validFromField validToField labelField iconUrl version updatedDate pathCost managedById folderId } }`, object{"id": id}, "data", "dataset")
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	return unpackObject(obj.(object), &objectDataset{}, ot.TypeName()), nil
}
