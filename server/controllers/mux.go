package controllers

import (
	"context"
	"github.com/muhtutorials/reminders_cli/server/models"
	"github.com/muhtutorials/reminders_cli/server/transport"
	"net/http"
	"regexp"
	"strings"
)

const paramsKey = "ps"

type ctxKey string

type urlParam struct {
	name     string
	regex    string
	value    string
	position int
}

type route struct {
	method  string
	path    string
	params  map[string]urlParam
	handler http.Handler
}

// populate populates a route based on the actual serving request
func (r *route) populate(req *http.Request) string {
	urlSlice := splitUrl(req.URL.Path)
	pathSlice := splitUrl(r.path)
	if len(pathSlice) != len(urlSlice) {
		return ""
	}
	for name, param := range r.params {
		regexParamValue := urlSlice[param.position]
		regex := regexp.MustCompile(param.regex)
		if regex.MatchString(regexParamValue) {
			param.value = regexParamValue
			r.params[name] = param
			pathSlice[param.position] = regexParamValue
		}
	}
	pathStr := "/" + strings.Join(pathSlice, "/")
	if req.URL.Path == pathStr {
		return r.method + pathStr
	}
	return ""
}

type RegexMux struct {
	routes    []*route
	routesMap map[string]*route
}

// Get registers an HTTP handler with GET method
func (m *RegexMux) Get(pattern string, handler http.Handler) {
	m.Handle(http.MethodGet, pattern, handler)
}

// Post registers an HTTP handler with POST method
func (m *RegexMux) Post(pattern string, handler http.Handler) {
	m.Handle(http.MethodPost, pattern, handler)
}

// Patch registers an HTTP handler with PATCH method
func (m *RegexMux) Patch(pattern string, handler http.Handler) {
	m.Handle(http.MethodPatch, pattern, handler)
}

// Put registers an HTTP handler with PUT method
func (m *RegexMux) Put(pattern string, handler http.Handler) {
	m.Handle(http.MethodPut, pattern, handler)
}

// Delete registers an HTTP handler with DELETE method
func (m *RegexMux) Delete(pattern string, handler http.Handler) {
	m.Handle(http.MethodDelete, pattern, handler)
}

func (m *RegexMux) Handle(method, pattern string, handler http.Handler) {
	params := m.getParams(pattern)
	r := &route{
		method:  method,
		path:    pattern,
		params:  params,
		handler: handler,
	}
	m.routes = append(m.routes, r)
}

func (m RegexMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.routesMap = map[string]*route{}
	for _, rt := range m.routes {
		key := rt.populate(r)
		m.routesMap[key] = rt
	}
	key := r.Method + r.URL.Path
	rt, ok := m.routesMap[key]
	if !ok {
		transport.SendError(w, models.NotFoundError{})
	}
	ctx := r.Context()
	if len(rt.params) != 0 {
		ctx = context.WithValue(ctx, ctxKey(paramsKey), rt.params)
	}
	rt.handler.ServeHTTP(w, r.WithContext(ctx))
}

// getParams retrieves a map of url params for a given url
func (m RegexMux) getParams(url string) map[string]urlParam {
	params := map[string]urlParam{}
	for _, v := range splitUrl(url) {
		urlPrm := m.parseParam(url, v)
		if urlPrm.name != "" {
			params[urlPrm.name] = urlPrm
		}
	}
	return params
}

// parseParam parses URL parameters given a {param}:Regex expression
func (m RegexMux) parseParam(url, regexParam string) urlParam {
	r := regexp.MustCompile("({[a-z]+}:)(.+)")
	matches := r.FindStringSubmatch(regexParam)
	// 1 - entire match: {[a-z]+}:.+
	// 2 - 1st group -> param name: {[a-z]+}:
	// 3 - 2nd group -> param regex: .+
	if len(matches) < 3 {
		return urlParam{}
	}
	replacer := strings.NewReplacer("{", "", "}", "", ":", "")
	// extract param name by removing "{"s and ":"
	name := replacer.Replace(matches[1])
	regex := matches[2]

	var position int
	for i, v := range splitUrl(url) {
		if v == matches[0] {
			position = i
		}
	}
	return urlParam{name: name, regex: regex, position: position}
}

// splitURL splits the request URL by "/" and returns a slice
func splitUrl(url string) []string {
	var result []string
	for _, part := range strings.Split(strings.TrimSpace(url), "/") {
		if strings.TrimSpace(part) != "" {
			result = append(result, part)
		}
	}
	return result
}
