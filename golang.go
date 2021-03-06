/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"

	"github.com/super-ast/gotranslate"
)

func encodeFromGo(r io.Reader, w io.Writer) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "in.go", r, 0)
	if err != nil {
		return err
	}
	a := gotranslate.NewAST(fset)
	ast.Walk(a, f)
	return json.NewEncoder(w).Encode(a.RootBlock)
}
