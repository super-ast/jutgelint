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
	if _, err := io.Copy(stdin, r); err != nil {
		return err
	}
	if err := stdin.Close(); err != nil {
		return err
	}
	return cmd.Wait()
}
