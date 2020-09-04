package ext

import (
    "encoding/json"
    "github.com/mattn/go-isatty"
    "github.com/nsf/jsondiff"
    "github.com/tidwall/gjson"
    "github.com/tidwall/sjson"
    "github.com/xsyr/goawk/interp"
    diff "github.com/yudai/gojsondiff"
    "github.com/yudai/gojsondiff/formatter"
    jaydiff "github.com/yazgazan/jaydiff/diff"
    "log"
    "os"
    "strings"
)

type jsonModule struct {
    funcs []*FuncDef
}

func newJsonModule() Module {
    var m = &jsonModule{}
    m.define()
    return m
}

func (m *jsonModule) Desc() string {
    return "json"
}
func (m *jsonModule) Funcs() []*FuncDef {
    return m.funcs
}

func (m *jsonModule)SetRuntime(rt *interp.Runtime) {

}

func (m *jsonModule) define() {
    m.funcs = []*FuncDef{
        {
            Desc: "__json_set_str(str, path, value) string. see https://github.com/tidwall/sjson",
            Func: func(s, p string, v string) string {
                val, err := sjson.Set(s, p, v)
                if err != nil {
                    log.Printf("sjson.Set() failed: %v", err)
                    return ""
                }
                return val
            },
        },
        {
            Desc: "__json_set_num(str, path, value) string. see https://github.com/tidwall/sjson",
            Func: func(s, p string, v float64) string {
                val, err := sjson.Set(s, p, v)
                if err != nil {
                    log.Printf("sjson.Set() failed: %v", err)
                    return ""
                }
                return val
            },
        },
        {
            Desc: "__json_diff(a,b) string. diff a and b. return empty if equal.",
            Func: func(a,b string) string {
                opt := jsondiff.DefaultConsoleOptions()
                if !isatty.IsTerminal(os.Stdout.Fd()) {
                    opt = jsondiff.Options{}
                }
                res, diff := jsondiff.Compare([]byte(a), []byte(b), &opt)
                if res == jsondiff.FullMatch {
                    return ""
                }
                return diff
            },
        },
        {
            Desc: "__json_diff2(a,b) string. diff a and b. return empty if equal.",
            Func: func(a,b string) string {
                differ := diff.New()
                d, err := differ.Compare([]byte(a), []byte(b))
                if err != nil {
                    log.Printf("differ.Compare() failed: %s", err)
                    return ""
                }
                if !d.Modified() {
                    return ""
                }
                var aJson map[string]interface{}
                json.Unmarshal([]byte(a), &aJson)

                config := formatter.AsciiFormatterConfig{
                    ShowArrayIndex: true,
                    Coloring: isatty.IsTerminal(os.Stdout.Fd()),
                }

                formatter := formatter.NewAsciiFormatter(aJson, config)
                diffString, err := formatter.Format(d)
                if err != nil {
                    log.Printf("formatter.Format() failed: %s", err)
                    return ""
                }
                return diffString
            },
        },
        {
            Desc: "__json_diffonly(a,b) string. diff a and b. return empty if equal.",
            Func: func(a,b string) string {
                var aJson, bJson map[string]interface{}
                if err := json.Unmarshal([]byte(a), &aJson); err != nil {
                    log.Printf("%s is invalid json, err: %s", a, err)
                    return ""
                }
                if err := json.Unmarshal([]byte(b), &bJson); err != nil {
                    log.Printf("%s is invalid json, err: %s", b, err)
                    return ""
                }
                d, err := jaydiff.Diff(aJson, bJson)
                if err != nil {
                    log.Printf("jaydiff.Diff() failed: %s", err)
                    return ""
                }
                output := jaydiff.Output{
                    Indent: "    ",
                    ShowTypes:true,
                    Colorized:isatty.IsTerminal(os.Stdout.Fd()),
                }
                ss, err := jaydiff.Report(d, output)
                if err != nil {
                    log.Printf("jaydiff.Report() failed: %s", err)
                    return ""
                }
                return strings.Join(ss, "\n")
            },
        },
        {
            Desc: "__json_get(str, path) string. see https://github.com/tidwall/gjson",
            Func: func(s, p string) string {
                return gjson.Get(s, p).String()
            },
        },
        {
            Desc: "__json_get_many(str, args....) []string. see https://github.com/tidwall/gjson",
            Func: func(s string, path... string) []string {
                out := make([]string, len(path))
                for i, r := range gjson.GetMany(s, path...) {
                    out[i] = r.String()
                }
                return out
            },
        },
        {
            Desc: "__json_typeof(str, path) string. see https://github.com/tidwall/gjson",
            Func: func(s, p string) string {
                return gjson.Get(s, p).Type.String()
            },
        },
        {
            Desc: "__json_isarray(str, path) bool. 0=no, 1=yes. see https://github.com/tidwall/gjson",
            Func: func(s, p string) bool {
                return gjson.Get(s, p).IsArray()
            },
        },
        {
            Desc: "__json_isobject(str, path) bool. 0=no, 1=yes. see https://github.com/tidwall/gjson",
            Func: func(s, p string) bool {
                return gjson.Get(s, p).IsObject()
            },
        },
        {
            Desc: "__json_isvalid(str) bool. 0=no, 1=yes. see https://github.com/tidwall/gjson",
            Func: func(s string) bool {
                return gjson.Valid(s)
            },
        },
        {
            Desc: "__json_encode(map[string]string) string.",
            Func: func(m map[string]string) string {
                res, _ := json.Marshal(m)
                return string(res)
            },
        },
    }

    for _, fn := range m.funcs {
        fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
    }
}