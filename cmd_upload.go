package main

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/exp/maps"
)

var (
	flagsUpload          *pflag.FlagSet
	flagUploadAsFilename string
	flagUploadDocumentId string
)

var ErrUploadUsage = ObserveError{Msg: "usage: observe upload <document usage> <document name>"}
var ErrUnsupportedType = ObserveError{Msg: "the document usage kind is not supported"}
var ErrNotTextFile = ObserveError{Msg: "the document is not a text file"}
var ErrFileNotReadable = ObserveError{Msg: "the file is not readable"}

func init() {
	flagsUpload = pflag.NewFlagSet("upload", pflag.ContinueOnError)
	flagsUpload.StringVarP(&flagUploadAsFilename, "as-filename", "f", "", "use this as uploaded name instead of the source file path")
	flagsUpload.StringVarP(&flagUploadDocumentId, "document-id", "d", "", "replace this particular document id, rather than matching on name")
	RegisterCommand(&Command{
		Name:  "upload",
		Help:  "Put (or overwrite) a document of some sort, identified by filename or id.",
		Flags: flagsUpload,
		Func:  cmdUpload,
	})
}

// 'upload' and 'document' are somewhat intertwined
func cmdUpload(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ErrUploadUsage
	}
	kind, has := docTypes[fa.args[1]]
	if !has {
		return ObserveError{Msg: "supported document usage kinds are: " + strings.Join(sorted(maps.Keys(docTypes)), ", ")}.WithInner(ErrUnsupportedType)
	}
	data, err := fa.fs.ReadFile(fa.args[2])
	if err != nil {
		return ErrFileNotReadable.WithInner(err)
	}
	mimetype, err := kind.SniffMimetype(data)
	if err != nil {
		return ErrNotTextFile.WithInner(err)
	}
	// is this a "create" or a "replace"?
	asName := fa.args[2]
	if flagUploadAsFilename != "" {
		asName = flagUploadAsFilename
	}
	fa.op.Debug("as_name=%s\n", asName)
	prevId := flagUploadDocumentId
	if prevId == "" {
		prevId = determinePreviousDocumentId(fa.cfg, fa.op, fa.hc, asName)
	}
	fa.op.Debug("prev_id=%s\n", prevId)
	var obj object
	if prevId == "" {
		obj, err = Query(fa.hc).Config(fa.cfg).Output(fa.op).Path("/v1/document/").Args(map[string]string{"name": asName, "usage": kind.Kind()}).Header(headers("content-type", mimetype)).Body(bytes.NewReader(data)).PropMap(propertyMapDocument).Post()
	} else {
		obj, err = Query(fa.hc).Config(fa.cfg).Output(fa.op).Path("/v1/document/" + url.QueryEscape(prevId)).Args(map[string]string{"name": asName, "usage": kind.Kind()}).Header(headers("content-type", mimetype)).Body(bytes.NewReader(data)).PropMap(propertyMapDocument).Put()
	}
	if err != nil {
		return err
	}
	keys := sorted(maps.Keys(obj))
	for _, k := range keys {
		fmt.Fprintf(fa.op, "%s: %v\n", k, obj[k])
	}
	return nil
}
