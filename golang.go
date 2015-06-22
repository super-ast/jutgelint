/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"

	"github.com/mvdan/superast"
)

func encodeJsonFromGoCode(i io.Reader, filename string, w io.Writer) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, i, 0)
	if err != nil {
		return err
	}
	a := superast.NewAST(fset)
	ast.Walk(a, f)

	enc := json.NewEncoder(w)
	return enc.Encode(a.RootBlock)
}
