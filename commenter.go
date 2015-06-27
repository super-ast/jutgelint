/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

func CheckAndCommentCode(lang Lang, r io.Reader, w io.Writer) error {
	code, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("could not read code: %v")
	}
	var json bytes.Buffer
	if err := EncodeJsonAST(lang, bytes.NewReader(code), &json); err != nil {
		return fmt.Errorf("could not translate code: %v")
	}
	warns, err := RunChecker(&json, CheckAll)
	if err != nil {
		return fmt.Errorf("could not check code: %v")
	}
	if err := CommentCode(warns, bytes.NewReader(code), w); err != nil {
		return fmt.Errorf("could not comment code: %v")
	}
	return nil
}

func CommentCode(warns Warnings, r io.Reader, w io.Writer) error {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	byLine := make([][]Warning, len(lines))
	for c := range warns {
		for _, warn := range warns[c] {
			// Assuming they start by 1 in our json AST
			ln := warn.Line - 1
			if ln < 0 || ln >= len(lines) {
				return errors.New("incorrect number of lines")
			}
			l := &byLine[ln]
			*l = append(*l, warn)
		}
	}
	for i, l := range lines {
		lineWarns := byLine[i]
		fmt.Fprintf(w, l)
		for j, warn := range lineWarns {
			if j == 0 {
				fmt.Fprintf(w, " // ")
			} else {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, warn.Long)
		}
		fmt.Fprintf(w, "\n")
	}
	return nil
}
