/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

func CommentCode(warns Warnings, r io.Reader, w io.Writer) error {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	byLine := make([][]Warning, len(lines))
	for c := range warns {
		for _, warn := range warns[c] {
			if warn.Line < 0 || warn.Line > len(lines) {
				return errors.New("incorrect number of lines")
			}
			l := &byLine[warn.Line]
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
