// +build gofuzz

package fuzz

import (
	"github.com/xsyr/goawk/parser"
)

func Fuzz(data []byte) int {
	if _, err := parser.ParseProgram(data, nil); err != nil {
		return 0
	}

	return 1
}
