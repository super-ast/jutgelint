/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"io"
	"os/exec"
)

func encodeFromCpp(r io.Reader, w io.Writer) error {
	cmd := exec.Command("superast-cpp")
	cmd.Stdin = r
	cmd.Stdout = w
	return cmd.Run()
}
