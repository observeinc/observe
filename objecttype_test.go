package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func estr(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}

func TestUnpackProppath(t *testing.T) {
	for i, tc := range []struct {
		input  any
		path   proppath
		output any
		error  string
		caller string
	}{
		{object{"key": "value"}, mkpath("key"), "value", "", getCaller(0)},
		{object{"key": "value"}, mkpath("value"), nil, `path "value": the 'data' is not an object`, getCaller(0)},
		{object{"deep": object{"thing": "here"}}, mkpath("deep.thing"), "here", "", getCaller(0)},
		{object{"key": array{}}, mkpath("key.value"), nil, `path "key.value": the 'data' is not an object`, getCaller(0)},
		{array{"key"}, mkpath("key"), nil, `path "key": the 'data' is not an object`, getCaller(0)},
		{nil, mkpath("key"), nil, `path "key": the 'data' is not an object`, getCaller(0)},
		{object{"key": object{"second": nil}}, mkpath("key.second"), nil, "", getCaller(0)},
	} {
		a, e := unpackProppath(tc.input, tc.path)
		if diff := cmp.Diff(a, tc.output); diff != "" {
			fmt.Fprintf(os.Stderr, "%s: case %d: unexpected diff:\n%s", tc.caller, i, diff)
			t.Fail()
		}
		if diff := cmp.Diff(estr(e), tc.error); diff != "" {
			fmt.Fprintf(os.Stderr, "%s: case %d: unexpected error:\n%s", tc.caller, i, diff)
			t.Fail()
		}
	}
}

func TestPropmapObject(t *testing.T) {
	pm := PropertyMap{
		"id":    mkpath("meta.id"),
		"thing": mkpath("config.value"),
		"score": mkpath("state.score"),
	}
	res, err := propmapObject(object{
		"type": "ignored",
		"meta": object{
			"id":   "123456",
			"name": "somename",
		},
		"config": object{
			"enabled": true,
			"value":   "booya",
		},
		"state": object{
			"score": 3.0,
		},
	}, pm)
	if err != nil {
		t.Fatal("error:", err)
	}
	if diff := cmp.Diff(res, object{
		"id":    "123456",
		"thing": "booya",
		"score": 3.0,
	}); diff != "" {
		t.Fatalf("unexpected diff:\n%s", diff)
	}
}
