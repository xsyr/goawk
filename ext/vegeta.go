package ext

import (
    "github.com/xsyr/goawk/interp"
    "strings"
    "encoding/json"
    "encoding/base64"
)

type vegetaModule struct {
    funcs []*FuncDef
}


func newVegetaModule() Module {
    var m = &vegetaModule{}
    m.define()
    return m
}

func (m *vegetaModule) Desc() string {
    return "vegeta"
}
func (m *vegetaModule) Funcs() []*FuncDef {
    return m.funcs
}

func (m *vegetaModule)SetRuntime(rt *interp.Runtime) {

}

func (m *vegetaModule) define() {
    m.funcs = []*FuncDef{
        {
            Desc: "__vegeta(method, url, body string, headers map[string]string) string.",
            Func: func(method, url, body string, headers map[string]string) string {
                hs := make(map[string][]string)
                for k, v := range headers {
                    hs[k] = []string{v}
                }
                m := map[string]interface{}{
                    "method": method,
                    "url": url,
                    "header": hs,
                }
                if method == "POST" {
                    m["body"] = base64.StdEncoding.EncodeToString([]byte(body))
                }
                data, err := json.Marshal(m)
                if err != nil {
                    panic(err)
                }
                return string(data)
            },
        },
    }

    for _, fn := range m.funcs {
        fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
    }
}