package ext

import (
	"encoding/json"
	"github.com/imroc/req"
	"github.com/xsyr/goawk/interp"
	"net/http"
	"strconv"
	"strings"
)

type httpModule struct {
	funcs []*FuncDef
}

func newHttpModule() Module {
	var m = &httpModule{}
	m.define()
	return m
}

func (m *httpModule) Desc() string {
	return `
http - e.g.
    {
        params["k1"] = "v1";
        params["k2"] = "v2";
        headers["DONTCLOSE"]="OK";
        resp[] = __http_get("http://www.baidu.com", params[], headers[]);
        print resp["code"];
        print resp["headers"];
        print resp["body"]

        __http_post("http://www.baidu.com", "this is body", headers[]);
    }

`
}
func (m *httpModule) Funcs() []*FuncDef {
	return m.funcs
}

func (m *httpModule)SetRuntime(rt *interp.Runtime) {

}

func headerToJson(hs http.Header) string {
	hj := make(map[string]string)
	for k, v := range hs {
		if len(v) > 0 {
			hj[k] = v[0]
		} else {
			hj[k] = ""
		}
	}
	r, _ := json.Marshal(hj)
	return string(r)
}

func (m *httpModule) define() {
	m.funcs = []*FuncDef{
		{
			Desc: "__http_get(url string, params, headers map[string]string) map[string]string. include err, code, header, body.",
			Func: func(s string, params, headers map[string]string) map[string]string {
				var ps  req.QueryParam
				if len(params) > 0 {
					ps = make(req.QueryParam)
					for k, v := range params {
						ps[k] = v
					}
				}
				r := make(map[string]string)
				res, err := req.Get(s, ps, req.HeaderFromStruct(headers))
				if err != nil {
					r["err"] = err.Error()
					return r
				}
				resp := res.Response()
				r["code"] = strconv.Itoa(resp.StatusCode)
				r["header"] = headerToJson(resp.Header)
				r["body"] = res.String()
				return r
			},
		},
		{
			Desc: "__http_post(url, body string, headers  map[string]string) map[string]string. include err, code, header, body.",
			Func: func(s string, body string, headers map[string]string) map[string]string {
				res, err := req.Post(s, body, req.HeaderFromStruct(headers))
				r := make(map[string]string)
				if err != nil {
					r["err"] = err.Error()
					return r
				}
				resp := res.Response()
				r["code"] = strconv.Itoa(resp.StatusCode)
				r["header"] = headerToJson(resp.Header)
				r["body"] = res.String()
				return r
			},
		},
		{
			Desc: "__http_debug(0/1).",
			Func: func(on int)  {
				req.Debug = on == 1
			},
		},
	}

	for _, fn := range m.funcs {
		fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
	}
}