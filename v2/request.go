package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func RequestPOST[Req, Resp any](cfg *Config, op Output, hc *http.Client, path string, req *Req, resp *Resp) (error, int) {
	body, err := json.Marshal(req)
	if err != nil {
		return err, -1
	}
	url := ClusterUrl(cfg, path)
	op.Debug("url=%s\n", url)
	op.Debug("payload=%s\n", string(body))
	// TODO: authorization header is needed for things other than login
	hresp, err := hc.Post(url.String(), "application/json", bytes.NewBuffer(body))
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

func RequestPOSTWithBodyOutput[Req any](cfg *Config, op Output, hc *http.Client, path string, req *Req, headers map[string]string, resp io.Writer) (error, int) {
	body, err := json.Marshal(req)
	if err != nil {
		return err, -1
	}
	url := ClusterUrl(cfg, path)
	op.Debug("url=%s\n", url)
	op.Debug("payload=%s\n", string(body))
	request, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(body))
	if err != nil {
		op.Debug("error=%s\n", err)
		return NewObserveError(err, "Client error"), -1
	}
	request.Header.Add("Content-Type", "application/json")
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	for h, v := range request.Header {
		op.Debug("%s=%s\n", h, strings.Join(v, "\n  "))
	}
	hresp, err := hc.Do(request)
	if err != nil {
		op.Debug("error=%s\n", err)
		return NewObserveError(err, "Network error"), -1
	}
	defer hresp.Body.Close()
	op.Debug("status=%d\n", hresp.StatusCode)
	if hresp.StatusCode > 299 {
		var msg map[string]any
		if nil == json.NewDecoder(hresp.Body).Decode(&msg) {
			if m, ok := msg["message"]; ok {
				return NewObserveError(nil, "%s", m), hresp.StatusCode
			}
		}
		return NewObserveError(nil, "HTTP error: %d", hresp.StatusCode), hresp.StatusCode
	}
	for h, v := range hresp.Header {
		op.Debug("%s=%s\n", h, strings.Join(v, "\n  "))
	}
	written, err := io.Copy(resp, hresp.Body)
	op.Debug("bytes_written=%d\n", written)
	if err != nil {
		return NewObserveError(err, "Error writing response"), hresp.StatusCode
	}
	return err, hresp.StatusCode
}

func RequestGET[Resp any](cfg *Config, op Output, hc *http.Client, path string, args map[string]string, resp *Resp) (error, int) {
	url := ClusterUrl(cfg, path)
	if len(args) > 0 {
		q := url.Query()
		for k, v := range args {
			q.Set(k, v)
		}
		url.RawQuery = q.Encode()
	}
	op.Debug("url=%s\n", url)
	// TODO: authorization header is needed for things other than login
	hresp, err := hc.Get(url.String())
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

var dottedQuadRex = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+(:\d+)?$`)

func ClusterUrl(cfg *Config, path string) *url.URL {
	protocol := "https://"
	cidStr := cfg.CustomerIdStr + "."
	// hack it to allow sandboxes to work easily
	if strings.HasSuffix(cfg.ClusterStr, ":4444") {
		protocol = "http://"
	}
	if dottedQuadRex.MatchString(cfg.ClusterStr) {
		cidStr = ""
		protocol = "http://"
	}
	str := fmt.Sprintf("%s%s%s%s", protocol, cidStr, cfg.ClusterStr, path)
	ret, err := url.Parse(str)
	if err != nil {
		panic(fmt.Sprintf("Somehow, a bad URL was constructed: %q: %s", str, err))
	}
	return ret
}
