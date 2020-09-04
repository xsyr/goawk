package ext

import (
    "github.com/xsyr/goawk/interp"
)

// FuncDef extend function definition
type FuncDef struct {
    Name string
    Desc string
    Func interface{}
}

type Module interface {
    Desc() string
    Funcs() []*FuncDef
    SetRuntime(rt *interp.Runtime)
}

var Modules []Module

func init() {
    Modules = []Module {
        newGorModule(),
        newJsonModule(),
        newMsgpackModule(),
        newRedisModule(),
        newUrlModule(),
        newCsvModule(),
        newHttpModule(),
    }
}