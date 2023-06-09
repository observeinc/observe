package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

func RequestPOST[Req, Resp any](cfg *Config, op Output, hc *http.Client, path string, req Req, resp *Resp) (error, int) {
	body, err := json.Marshal(req)
	if err != nil {
		return err, -1
	}
	hresp, err := requestCommon(cfg, op, hc, "POST", path, bytes.NewBuffer(body), nil, nil)
	if err != nil {
		op.Debug("error=%s\n", err)
		return err, -1
	}
	defer hresp.Body.Close()
	data, err := io.ReadAll(hresp.Body)
	if err != nil {
		op.Debug("error=%s\n", err)
		return err, hresp.StatusCode
	}
	op.Debug("response=%s\n", string(data))
	return json.Unmarshal(data, resp), hresp.StatusCode
}

func RequestPOSTWithBodyOutput[Req any](cfg *Config, op Output, hc *http.Client, path string, req Req, headers map[string]string, resp io.Writer) (error, int) {
	body, err := json.Marshal(req)
	if err != nil {
		return err, -1
	}
	var hresp *http.Response
	hresp, err = requestCommon(cfg, op, hc, "POST", path, bytes.NewBuffer(body), nil, headers)
	if err != nil {
		op.Debug("error=%s\n", err)
		return NewObserveError(err, "Network error"), -1
	}
	defer hresp.Body.Close()
	op.Debug("status=%d\n", hresp.StatusCode)
	logHeaders(op, hresp.Header)
	if hresp.StatusCode > 299 {
		var msg map[string]any
		var data []byte
		data, err = io.ReadAll(hresp.Body)
		if err == nil {
			op.Debug("body=%s\n", data)
			if nil == json.Unmarshal(data, &msg) {
				if m, ok := msg["message"]; ok {
					return NewObserveError(nil, "%s", m), hresp.StatusCode
				}
			}
		}
		return NewObserveError(nil, "HTTP error: %d", hresp.StatusCode), hresp.StatusCode
	}
	var written int64
	written, err = io.Copy(resp, hresp.Body)
	op.Debug("bytes_written=%d\n", written)
	if err != nil {
		return NewObserveError(err, "Error writing response"), hresp.StatusCode
	}
	return err, hresp.StatusCode
}

func RequestGET[Resp any](cfg *Config, op Output, hc *http.Client, path string, args map[string]string, resp *Resp) (error, int) {
	hresp, err := requestCommon(cfg, op, hc, "GET", path, nil, args, nil)
	if err != nil {
		return err, -1
	}
	defer hresp.Body.Close()
	data, err := io.ReadAll(hresp.Body)
	if err != nil {
		return err, hresp.StatusCode
	}
	op.Debug("response=%s\n", string(data))
	return json.Unmarshal(data, resp), hresp.StatusCode
}

func requestCommon(cfg *Config, op Output, hc *http.Client, verb string, path string, body io.Reader, args map[string]string, headers map[string]string) (*http.Response, error) {
	url := SiteUrl(cfg, path)
	if len(args) > 0 {
		q := url.Query()
		for k, v := range args {
			q.Set(k, v)
		}
		url.RawQuery = q.Encode()
	}
	op.Debug("url=%s\n", url)
	request, err := http.NewRequest(verb, url.String(), body)
	if err != nil {
		op.Debug("error=%s\n", err)
		return nil, NewObserveError(err, "Client error")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", fmt.Sprintf("observe/%s (%s) g=%s", strings.TrimSpace(ReleaseVersion), runtime.GOOS, strings.TrimSpace(GitCommit)))
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	logHeaders(op, request.Header)
	for h, v := range request.Header {
		op.Debug("%s=%s\n", h, strings.Join(v, "\n  "))
	}
	return hc.Do(request)
}

var dottedQuadRex = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+(:\d+)?$`)

func SiteUrl(cfg *Config, path string) *url.URL {
	protocol := "https://"
	cidStr := cfg.CustomerIdStr + "."
	// hack it to allow sandboxes to work easily
	if strings.HasSuffix(cfg.SiteStr, ":4444") {
		protocol = "http://"
	}
	if dottedQuadRex.MatchString(cfg.SiteStr) {
		cidStr = ""
		protocol = "http://"
	}
	str := fmt.Sprintf("%s%s%s%s", protocol, cidStr, cfg.SiteStr, path)
	ret, err := url.Parse(str)
	if err != nil {
		panic(fmt.Sprintf("Somehow, a bad URL was constructed: %q: %s", str, err))
	}
	return ret
}

func logHeaders(op Output, headers http.Header) {
	var keys []string
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		op.Debug("%s=%s\n", k, strings.Join(headers[k], "\n  "))
	}
}
