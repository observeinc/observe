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

func RequestPOST[Req, Resp any](cfg *Config, op Output, hc httpClient, path string, req Req, resp *Resp, headers http.Header) (error, int) {
	body, err := json.Marshal(req)
	if err != nil {
		return err, -1
	}
	hresp, err := Query(hc).Config(cfg).Output(op).Path(path).Body(bytes.NewBuffer(body)).Header(headers).requestCommon("POST")
	if err != nil {
		op.Debug("error=%s\n", err)
		return err, -1
	}
	defer hresp.Body.Close()
	if hresp.StatusCode > 299 {
		return HttpStatusError(op, path, hresp), hresp.StatusCode
	}
	data, err := io.ReadAll(hresp.Body)
	if err != nil {
		op.Debug("error=%s\n", err)
		return err, hresp.StatusCode
	}
	op.Debug("response=%s\n", string(data))
	return json.Unmarshal(data, resp), hresp.StatusCode
}

func RequestPOSTWithBodyOutput[Req any](cfg *Config, op Output, hc httpClient, path string, req Req, headers http.Header, resp io.Writer) (error, int) {
	body, err := json.Marshal(req)
	if err != nil {
		return err, -1
	}
	var hresp *http.Response
	hresp, err = Query(hc).Config(cfg).Output(op).Path(path).Body(bytes.NewBuffer(body)).Header(headers).requestCommon("POST")
	if err != nil {
		op.Debug("error=%s\n", err)
		return NewObserveError(err, "Network error"), -1
	}
	defer hresp.Body.Close()
	if hresp.StatusCode > 299 {
		return HttpStatusError(op, path, hresp), hresp.StatusCode
	}
	var written int64
	written, err = io.Copy(resp, hresp.Body)
	op.Debug("bytes_written=%d\n", written)
	if err != nil {
		return NewObserveError(err, "Error writing response"), hresp.StatusCode
	}
	if sfqid := hresp.Header.Get("X-Observe-Sfqid"); sfqid != "" {
		op.Debug("X-Observe-Sfqid=%s\n", sfqid)
	}
	return err, hresp.StatusCode
}

func RequestGET[Resp any](cfg *Config, op Output, hc httpClient, path string, args map[string]string, headers http.Header, resp *Resp) (error, int) {
	hresp, err := Query(hc).Config(cfg).Output(op).Path(path).Args(args).Header(headers).requestCommon("GET")
	if err != nil {
		return err, -1
	}
	defer hresp.Body.Close()
	if hresp.StatusCode > 299 {
		return HttpStatusError(op, path, hresp), hresp.StatusCode
	}
	data, err := io.ReadAll(hresp.Body)
	if err != nil {
		return err, hresp.StatusCode
	}
	op.Debug("response=%s\n", string(data))
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return dec.Decode(resp), hresp.StatusCode
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

func HttpStatusError(op Output, url string, hresp *http.Response) error {
	var msg map[string]any
	data, err := io.ReadAll(hresp.Body)
	if err == nil {
		op.Debug("body=%s\n", data)
		err = json.Unmarshal(data, &msg)
	}
	if err == nil {
		if m, ok := msg["message"]; ok {
			return NewObserveError(nil, "%s", m)
		} else if m, ok := msg["errors"]; ok {
			return NewObserveError(nil, "%v", m)
		} else if m, ok := msg["error"]; ok {
			return NewObserveError(nil, "%s", m)
		}
	}
	return NewObserveError(nil, "%s: HTTP error: %d", url, hresp.StatusCode)
}

type ApiResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type pendingQuery struct {
	hc      httpClient
	cfg     *Config
	op      Output
	path    string
	args    map[string]string
	header  http.Header
	body    io.Reader
	propmap PropertyMap
}

func Query(hc httpClient) *pendingQuery {
	return &pendingQuery{
		hc: hc,
	}
}

func (p *pendingQuery) Config(cfg *Config) *pendingQuery {
	p.cfg = cfg
	return p
}

func (p *pendingQuery) Output(op Output) *pendingQuery {
	p.op = op
	return p
}

func (p *pendingQuery) Path(pa string) *pendingQuery {
	p.path = pa
	return p
}

func (p *pendingQuery) Args(args map[string]string) *pendingQuery {
	p.args = args
	return p
}

func (p *pendingQuery) Header(h http.Header) *pendingQuery {
	p.header = h
	return p
}

func (p *pendingQuery) Body(d io.Reader) *pendingQuery {
	p.body = d
	return p
}

func (p *pendingQuery) PropMap(pm PropertyMap) *pendingQuery {
	p.propmap = pm
	return p
}

func (p *pendingQuery) GetList() (array, error) {
	p.verifyWithProps()
	if p.body != nil {
		panic("body is not used in GetList")
	}
	var ar ApiResponse
	if err := p.queryDecode("GET", &ar); err != nil {
		return nil, err
	}
	if ar.Data == nil {
		return nil, ErrNotAnArray
	}
	in, is := ar.Data.(array)
	if !is {
		return nil, ErrNotAnArray
	}
	ret := make(array, len(in))
	for i := range in {
		var err error
		ret[i], err = propmapObject(in[i], p.propmap)
		if err != nil {
			return nil, NewObserveError(err, "%s: item %d", p.path, i)
		}
	}
	return ret, nil

}

func (p *pendingQuery) Get() (object, error) {
	p.verifyWithProps()
	if p.body != nil {
		panic("body is not used in GetList")
	}
	var ar ApiResponse
	if err := p.queryDecode("GET", &ar); err != nil {
		return nil, err
	}
	if ar.Data == nil {
		return nil, ErrNotAnObject
	}
	return propmapObject(ar.Data, p.propmap)
}

func (p *pendingQuery) Post() (object, error) {
	return p.putPostQuery("POST")
}

func (p *pendingQuery) Put() (object, error) {
	return p.putPostQuery("PUT")
}

func (p *pendingQuery) Delete() error {
	p.verifyBase()
	if p.body != nil {
		panic("body is not used in Delete")
	}
	var ar ApiResponse
	if err := p.queryDecode("DELETE", &ar); err != nil {
		return err
	}
	return nil
}

func (p *pendingQuery) putPostQuery(verb string) (object, error) {
	p.verifyWithBody()
	var ar ApiResponse
	if err := p.queryDecode(verb, &ar); err != nil {
		return nil, err
	}
	if ar.Data == nil {
		return nil, ErrNotAnObject
	}
	return propmapObject(ar.Data, p.propmap)
}

func (p *pendingQuery) queryDecode(verb string, ar *ApiResponse) error {
	hresp, err := p.queryInner(verb)
	if err != nil {
		return err
	}
	defer hresp.Body.Close()
	dec := json.NewDecoder(hresp.Body)
	if err = dec.Decode(ar); err != nil {
		return NewObserveError(err, "%s: resonse decode", p.path)
	}
	if !ar.Ok {
		return NewObserveError(nil, "%s: response not ok", p.path)
	}
	return nil
}

func (p *pendingQuery) queryInner(verb string) (*http.Response, error) {
	hresp, err := p.requestCommon(verb)
	if err != nil {
		return nil, NewObserveError(err, "%s: request", p.path)
	}
	if hresp.StatusCode > 299 {
		return nil, HttpStatusError(p.op, p.path, hresp)
	}
	return hresp, nil
}

func (p *pendingQuery) requestCommon(verb string) (*http.Response, error) {
	url := SiteUrl(p.cfg, p.path)
	if len(p.args) > 0 {
		q := url.Query()
		for k, v := range p.args {
			q.Set(k, v)
		}
		url.RawQuery = q.Encode()
	}
	p.op.Debug("url=%s\n", url)
	request, err := http.NewRequest(verb, url.String(), p.body)
	if err != nil {
		p.op.Debug("error=%s\n", err)
		return nil, NewObserveError(err, "request error")
	}
	request.Header.Set("Host", p.cfg.SiteStr)
	request.Header.Set("User-Agent", fmt.Sprintf("observe/%s (%s) g=%s", strings.TrimSpace(ReleaseVersion), runtime.GOOS, strings.TrimSpace(GitCommit)))
	request.Header.Set("Authorization", p.cfg.AuthHeader())
	//	this is a reasonable default, but we can override it with the headers parameter
	request.Header.Set("Content-Type", "application/json")
	for k, v := range p.header {
		request.Header.Set(k, v[0])
	}
	logHeaders(p.op, request.Header)
	resp, err := p.hc.Do(request)
	if resp != nil {
		p.op.Debug("status=%d\n", resp.StatusCode)
		logHeaders(p.op, resp.Header)
	} else {
		p.op.Debug("error=%s\n", err)
	}
	return resp, err
}

func (p *pendingQuery) verifyBase() {
	if p.hc == nil {
		panic("missing httpClient")
	}
	if p.cfg == nil {
		panic("missing Config")
	}
	if p.op == nil {
		panic("missing Output")
	}
	if p.path == "" {
		panic("missing path")
	}
}

func (p *pendingQuery) verifyWithProps() {
	p.verifyBase()
	if p.propmap == nil {
		panic("missing PropertyMap")
	}
}

func (p *pendingQuery) verifyWithBody() {
	p.verifyWithProps()
	if p.body == nil {
		panic("missing body")
	}
	// headers are optional
}
