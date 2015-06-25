/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"io"
	"os/exec"
)

func encodeFromCpp(r io.Reader, w io.Writer) error {
	cmd := exec.Command("superast-cpp")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stdout = w
	if err := cmd.Start(); err != nil {
		return err
	}
	io.Copy(stdin, r)
	stdin.Close()
	return cmd.Wait()
}
