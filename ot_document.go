package main

import (
	"fmt"
	"net/url"
)

func init() {
	RegisterObjectType(ObjectTypeDocument, &objectDocument{})
}

type objectDocument struct {
	objectDocumentMeta
	objectDocumentConfig
	objectDocumentState
}

type objectDocumentMeta struct {
	Id string
}

type objectDocumentConfig struct {
	Name  string
	Usage string
}

type objectDocumentState struct {
	Mimetype string
	Size     int64
	Url      string
	auditedObject
}

var _ ObjectInstance = &objectDocument{}

type documentKind interface {
	Kind() string
	SniffMimetype(data []byte) (string, error)
}

var docTypes = map[string]documentKind{}

func addDocumenKind(dk documentKind) {
	if _, has := docTypes[dk.Kind()]; has {
		panic(fmt.Sprintf("duplicate document kind: %s", dk.Kind()))
	}
	docTypes[dk.Kind()] = dk
}
func (o *objectDocument) GetInfo() *ObjectInfo {
	return &ObjectInfo{
		Id:           o.Id,
		Name:         o.Name,
		Presentation: []string{o.Id, o.Usage, o.Name},
	}
}

func (o *objectDocument) GetValues() []PropertyInstance {
	props := ObjectTypeDocument.GetProperties()
	r := make([]PropertyInstance, len(props))
	for i, p := range props {
		r[i] = &propertyInstance{p, o}
	}
	return r
}

func (o *objectDocument) GetStore() object {
	return nil
}

func (o *objectDocument) PrintToYaml(op Output, otyp ObjectType, obj ObjectInstance) error {
	return printToYamlFromObjectInstance(op, otyp, obj)
}

type objectTypeDocument struct{}

var ObjectTypeDocument ObjectType = &objectTypeDocument{}

var propertyDescDocument = []PropertyDesc{
	{"id", PropertyTypeString, false, true, func(o any) any { return o.(*objectDocument).Id }, func(o any, v any) { o.(*objectDocument).Id = v.(string) }},
	{"name", PropertyTypeString, false, false, func(o any) any { return o.(*objectDocument).Name }, func(o any, v any) { o.(*objectDocument).Name = v.(string) }},
	{"usage", PropertyTypeString, false, false, func(o any) any { return o.(*objectDocument).Usage }, func(o any, v any) { o.(*objectDocument).Usage = v.(string) }},
	{"mimetype", PropertyTypeString, true, false, func(o any) any { return o.(*objectDocument).Mimetype }, func(o any, v any) { o.(*objectDocument).Mimetype = v.(string) }},
	{"size", PropertyTypeInteger, true, false, func(o any) any { return o.(*objectDocument).Size }, func(o any, v any) { o.(*objectDocument).Size = v.(int64) }},
	{"url", PropertyTypeString, true, false, func(o any) any { return o.(*objectDocument).Url }, func(o any, v any) { o.(*objectDocument).Url = v.(string) }},
	{"updatedDate", PropertyTypeString, true, false, func(o any) any { return o.(*objectDocument).UpdatedDate }, func(o any, v any) { o.(*objectDocument).UpdatedDate = v.(string) }},
}

// Currently mapping from meta/config/state tuple, to flat.
// But we really want the tuple to be the "real" format.
var propertyMapDocument = PropertyMap{
	"id":          mkpath("meta.id"),
	"name":        mkpath("config.name"),
	"usage":       mkpath("config.usage"),
	"mimetype":    mkpath("state.mimetype"),
	"size":        mkpath("state.size"),
	"url":         mkpath("state.url"),
	"updatedDate": mkpath("state.updatedDate"),
}

func (*objectTypeDocument) TypeName() string                { return "document" }
func (*objectTypeDocument) Help() string                    { return "an uploaded auxiliary document" }
func (*objectTypeDocument) CanList() bool                   { return true }
func (*objectTypeDocument) CanGet() bool                    { return true }
func (*objectTypeDocument) CanCreate() bool                 { return false }
func (*objectTypeDocument) CanUpdate() bool                 { return false }
func (*objectTypeDocument) CanDelete() bool                 { return true }
func (*objectTypeDocument) GetPresentationLabels() []string { return []string{"id", "usage", "name"} }
func (*objectTypeDocument) GetProperties() []PropertyDesc   { return propertyDescDocument }

func (ot *objectTypeDocument) List(cfg *Config, op Output, hc httpClient) ([]*ObjectInfo, error) {
	cu, err := Query(hc).Config(cfg).Output(op).Path("/v1/document/").PropMap(propertyMapDocument).GetList()
	if err != nil || cu == nil {
		return nil, err
	}
	var ret []*ObjectInfo
	idp := getpropdesc(ot, "id")
	usagep := getpropdesc(ot, "usage")
	namep := getpropdesc(ot, "name")
	for _, doc := range cu {
		ret = append(ret, unpackInfo(
			doc.(object),
			idp,
			namep,
			idp,
			usagep,
			namep,
		))
	}
	return ret, nil
}

func (ot *objectTypeDocument) Get(cfg *Config, op Output, hc httpClient, id string) (ObjectInstance, error) {
	obj, err := Query(hc).Config(cfg).Output(op).Path("/v1/document/" + url.QueryEscape(id)).PropMap(propertyMapDocument).Get()
	if err != nil || obj == nil {
		return nil, err
	}
	return unpackObject(obj, &objectDocument{}, ot.TypeName()), nil
}

func (ot *objectTypeDocument) Create(cfg *Config, op Output, hc httpClient, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeDocument) Update(cfg *Config, op Output, hc httpClient, id string, input object) (ObjectInstance, error) {
	return nil, nil
}

func (ot *objectTypeDocument) Delete(cfg *Config, op Output, hc httpClient, id string) error {
	return Query(hc).Config(cfg).Output(op).Path("/v1/document/" + url.QueryEscape(id)).Delete()
}

func determinePreviousDocumentId(cfg *Config, op Output, hc httpClient, asName string) string {
	lst, err := Query(hc).Config(cfg).Output(op).Path("/v1/document/").Args(map[string]string{"name": asName}).PropMap(propertyMapDocument).GetList()
	if err != nil || len(lst) < 1 {
		return ""
	}
	if val, is := lst[0].(object)["id"]; is {
		if str, is := val.(string); is {
			return str
		}
	}
	return ""
}
