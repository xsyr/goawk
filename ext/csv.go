package ext

import (
	"bytes"
	"encoding/csv"
	"github.com/xsyr/goawk/interp"
	"io"
	"log"
	"strings"
)

type csvModule struct {
	funcs []*FuncDef
}

func newCsvModule() Module {
	var m = &csvModule{}
	m.define()
	return m
}

func (m *csvModule) Desc() string {
	return "url"
}

func (m *csvModule) Funcs() []*FuncDef {
	return m.funcs
}

func (m *csvModule)SetRuntime(rt *interp.Runtime) {

}

func (m *csvModule) define() {
	m.funcs = []*FuncDef{
		{
			Desc: "__csv_decode(string) []string",
			Func: func(s string) []string {
				r := csv.NewReader(strings.NewReader(s))
				record, err := r.Read()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					log.Println(err)
				}
				return record
			},
		},
		{
			Desc: "__csv_encode([]string) string",
			Func: func(s []string) string {
				if len(s) == 0 { return "" }
				var buf bytes.Buffer
				w := csv.NewWriter(&buf)
				w.Write(s)
				w.Flush()
				return string(buf.Bytes()[:buf.Len()-1])
			},
		},
	}

	for _, fn := range m.funcs {
		fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
	}
}