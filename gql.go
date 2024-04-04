package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

var ErrNotAnArray = ObserveError{Msg: "the 'data' is not an array"}
var ErrNotAnObject = ObserveError{Msg: "the 'data' is not an object"}

func compileGqlQuery(q string, path ...string) compiledGqlQuery {
	return compiledGqlQuery{
		q:    q,
		path: path,
	}
}

type compiledGqlQuery struct {
	q     string
	path  []string
	remap remapPrepped
}

func (cq compiledGqlQuery) query(cfg *Config, op Output, hc httpClient, args object) (any, error) {
	// format arguments as GraphQL request object
	obj := object{"query": cq.q, "variables": args}
	buf := bytes.Buffer{}
	err, _ := RequestPOSTWithBodyOutput(cfg, op, hc, "/v1/meta", obj, headers("Authorization", cfg.AuthHeader()), &buf)
	data := buf.Bytes()
	if len(data) > 0 {
		op.Debug("payload=%s\n", data)
	}
	if err != nil {
		op.Debug("err=%s\n", err)
		return nil, err
	}
	return cq.unmarshalDecode(data)
}

func (cq compiledGqlQuery) unmarshalDecode(data []byte) (any, error) {
	var ret any
	err := json.Unmarshal(data, &ret)
	if err != nil {
		return nil, NewObserveError(err, "response JSON decode error")
	}
	retObj, is := ret.(object)
	if !is {
		return nil, NewObserveError(nil, "response is malformed: not an object")
	}
	if errs, has := retObj["errors"]; has {
		if y, is := errs.(array); is {
			return nil, NewObserveError(nil, "response JSON error: %v", y[0])
		}
		return nil, NewObserveError(nil, "response JSON error: %v", errs)
	}
	return cq.decodePath(ret)
}

func (cq compiledGqlQuery) decodePath(ret any) (any, error) {
	for i, key := range cq.path {
		if a, is := ret.(array); is {
			i64, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return nil, NewObserveError(nil, "response is array, key expected: %v", cq.path[:i+1])
			}
			if i64 < 0 || int(i64) >= len(a) {
				return nil, NewObserveError(nil, "response expected array with at least %d elements; got %d: %v", i64+1, len(a), cq.path[:i+1])
			}
			ret = a[int(i64)]
		} else if o, is := ret.(object); is {
			if ret, is = o[key]; !is {
				return nil, NewObserveError(nil, "response expected object with path: %v", cq.path[:i+1])
			}
		} else {
			return nil, NewObserveError(nil, "response JSON is not an object at path: %v", cq.path[:i+1])
		}
	}
	return cq.maybeRemap(ret)
}

func (cq compiledGqlQuery) maybeRemap(ret any) (any, error) {
	if cq.remap == nil {
		return ret, nil
	}
	if a, is := ret.(array); is {
		return cq.remapArray(a)
	}
	if o, is := ret.(object); is {
		return cq.remapObject(o)
	}
	return nil, NewObserveError(nil, "response JSON is not an object: %v", ret)
}

func (cq compiledGqlQuery) remapArray(a array) (any, error) {
	for i, ai := range a {
		if o, is := ai.(object); is {
			a[i], _ = cq.remapObject(o)
		} else if ai == nil {
			// do nothing
		} else {
			return nil, NewObserveError(nil, "response JSON at index %d is not an object: %v", i, o)
		}
	}
	return a, nil
}

func (cq compiledGqlQuery) remapObject(o object) (any, error) {
	// I don't know which of the paths are optional, so I just unpack what I can and sort it out later
	for dst, src := range cq.remap {
		val, err := unpackProppath(o, src)
		if err == nil {
			o[dst] = val
		}
	}
	for _, src := range cq.remap {
		delete(o, src[0])
	}
	return o, nil
}

type remap map[string]string
type remapPrepped map[string][]string

func prepRemap(r remap) remapPrepped {
	ret := remapPrepped{}
	for dst, src := range r {
		ret[dst] = strings.Split(src, ".")
	}
	return ret
}

func (cq compiledGqlQuery) WithRemap(r remap) compiledGqlQuery {
	cq.remap = prepRemap(r)
	return cq
}
