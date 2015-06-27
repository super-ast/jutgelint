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
	"sort"
)

func CheckAndCommentCode(lang Lang, r io.Reader, w io.Writer) error {
	code, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("could not read code: %v", err)
	}
	var json bytes.Buffer
	if err := EncodeJsonAST(lang, bytes.NewReader(code), &json); err != nil {
		return fmt.Errorf("could not translate code: %v", err)
	}
	warns, err := RunChecker(&json, CheckAll)
	if err != nil {
		return fmt.Errorf("could not check code: %v", err)
	}
	if err := CommentCode(lang, warns, bytes.NewReader(code), w); err != nil {
		return fmt.Errorf("could not comment code: %v", err)
	}
	return nil
}

type sortedWarns []Warning

func (sw sortedWarns) Len() int           { return len(sw) }
func (sw sortedWarns) Swap(i, j int)      { sw[i], sw[j] = sw[j], sw[i] }
func (sw sortedWarns) Less(i, j int) bool { return sw[i].Long < sw[j].Long }

func CommentCode(lang Lang, warns Warnings, r io.Reader, w io.Writer) error {
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
			lineWarns := &byLine[ln]
			*lineWarns = append(*lineWarns, warn)
		}
	}
	prefix := lang.InlineCommentPrefix()
	for i, line := range lines {
		lineWarns := byLine[i]
		sort.Sort(sortedWarns(lineWarns))
		fmt.Fprintf(w, line)
		for j, warn := range lineWarns {
			if j == 0 {
				fmt.Fprintf(w, prefix)
			} else {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, warn.Long)
		}
		fmt.Fprintf(w, "\n")
	}
	return nil
}
