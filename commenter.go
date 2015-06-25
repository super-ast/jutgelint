/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import "bytes"

func CommentCode(warns []Warning, lines []string) []string {
	byLine := make([][]Warning, len(lines))
	for _, w := range warns {
		l := &byLine[w.Line]
		*l = append(*l, w)
	}
	var comm []string
	for i, l := range lines {
		lineWarns := byLine[i]
		b := bytes.NewBufferString(l)
		b.WriteString(" // ")
		for i, w := range lineWarns {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(w.Long)
		}
		comm = append(comm, b.String())

	}
	return comm
}
