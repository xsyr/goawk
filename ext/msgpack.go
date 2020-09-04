package ext

import (
    "github.com/xsyr/goawk/interp"
    "log"
    "github.com/vmihailenco/msgpack"
    "github.com/json-iterator/go"
    "strings"
)

type msgpackModule struct {
    funcs []*FuncDef
}

func newMsgpackModule() Module {
    var m = &msgpackModule{}
    m.define()
    return m
}

func (m *msgpackModule) Desc() string {
    return "msgpack"
}

func (m *msgpackModule) Funcs() []*FuncDef {
    return m.funcs
}
func (m *msgpackModule)SetRuntime(rt *interp.Runtime) {

}

func (m *msgpackModule) define() {
    m.funcs = []*FuncDef{
        {
            Desc: "__msgp_encode(data) []byte",
            Func: func(data []byte) []byte {
                val, err := msgpack.Marshal(data)
                if err != nil {
                    log.Printf("msgpack.Marshal() failed: %v", err)
                    return nil
                }
                return val
            },
        },
        {
            Desc: "__msgp_decode(data) []byte",
            Func: func(data []byte) []byte {
                var val []byte
                err := msgpack.Unmarshal(data, &val)
                if err != nil {
                    log.Printf("msgpack.Unmarshal() failed: %v", err)
                    return nil
                }
                return val
            },
        },
        {
            Desc: "__msgp_decode2json(data) []byte",
            Func: func(data []byte) []byte {
                var val interface{}
                err := msgpack.Unmarshal(data, &val)
                if err != nil {
                    log.Printf("msgpack.Unmarshal() failed: %v", err)
                    return nil
                }
                var json = jsoniter.ConfigCompatibleWithStandardLibrary
                j, err := json.Marshal(val)
                if err != nil {
                    log.Printf("json.Marshal() failed: %v", err)
                    return nil
                }
                return j
            },
        },
    }

    for _, fn := range m.funcs {
        fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
    }
}