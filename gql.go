package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

func gqlQuery(cfg *Config, op Output, hc *http.Client, q string, args object, path ...string) (any, error) {
	// format arguments as GraphQL request object
	obj := object{"query": q, "variables": args}
	buf := bytes.Buffer{}
	err, _ := RequestPOSTWithBodyOutput(cfg, op, hc, "/v1/meta", obj, map[string]string{
		"Authorization": "Bearer " + cfg.CustomerIdStr + " " + cfg.AuthtokenStr,
	}, &buf)
	data := buf.Bytes()
	if len(data) > 0 {
		op.Debug("payload=%s\n", data)
	}
	if err != nil {
		op.Debug("err=%s\n", err)
		return nil, err
	}
	var ret any
	err = json.Unmarshal(data, &ret)
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
	for i, key := range path {
		if a, is := ret.(array); is {
			i64, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return nil, NewObserveError(nil, "response is array, key expected: %v", path[:i+1])
			}
			if i64 < 0 || int(i64) >= len(a) {
				return nil, NewObserveError(nil, "response expected array with at least %d elements; got %d: %v", i64+1, len(a), path[:i+1])
			}
			ret = a[int(i64)]
		} else if o, is := ret.(object); is {
			if ret, is = o[key]; !is {
				return nil, NewObserveError(nil, "response expected object with path: %v", path[:i+1])
			}
		} else {
			return nil, NewObserveError(nil, "response JSON is not an object at path: %v", path[:i+1])
		}
	}
	return ret, nil
}
