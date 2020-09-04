package ext

import (
    "github.com/xsyr/goawk/interp"
    "log"

    //"fmt"
    "net/url"
    "strings"
)

type urlModule struct {
    funcs []*FuncDef
}

func newUrlModule() Module {
    var m = &urlModule{}
    m.define()
    return m
}

func (m *urlModule) Desc() string {
    return "url"
}

func (m *urlModule) Funcs() []*FuncDef {
    return m.funcs
}

func (m *urlModule)SetRuntime(rt *interp.Runtime) {

}

func (m *urlModule) define() {
    m.funcs = []*FuncDef{
        {
            Desc: "__url_query_escape(string) string",
            Func: func(s string) string {
                return url.QueryEscape(s)
            },
        },
        {
            Desc: "__url_query_unescape(string) string",
            Func: func(s string) string {
                r, _ := url.QueryUnescape(s)
                return r
            },
        },
        {
            Desc: "__url_path_escape(string) string",
            Func: func(s string) string {
                return url.PathEscape(s)
            },
        },
        {
            Desc: "__url_path_unescape(string) string",
            Func: func(s string) string {
                r, _ := url.PathUnescape(s)
                return r
            },
        },
        {
            Desc: "__url_get_hostname(string) string",
            Func: func(s string) string {
                u, err := url.Parse(s)
                if err != nil {
                    return ""
                }
                return u.Hostname()
            },
        },
        {
            Desc: "__url_get_path(string) string",
            Func: func(s string) string {
                u, err := url.Parse(s)
                if err != nil {
                    return ""
                }
                return u.Path
            },
        },
        {
            Desc: "__url_get_param(url, name) string",
            Func: func(s, n string) string {
                u, err := url.Parse(s)
                if err != nil {
                    return ""
                }
                return u.Query().Get(n)
            },
        },
        {
            Desc: "__url_get_params(url) map[string]string",
            Func: func(s string) ( map[string]string) {
                u, err := url.Parse(s)
                if err != nil {
                    log.Printf("__url_get_params() can't parse '%s'", s)
                    return nil
                }
                m := make(map[string]string)
                for k, v := range u.Query() {
                    if len(v) > 0 {
                        m[k] = v[0]
                    } else {
                        m[k] = ""
                    }
                }
                return m
            },
        },
    }

    for _, fn := range m.funcs {
        fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
    }
}