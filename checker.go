/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"encoding/json"
	"io"
	"os/exec"
)

var checkOpts = []string{
	"--dead-assign",
	"--fors",
	"--local-decl",
	"--variable-init",
}

type Warnings map[string][]Warning

type Warning struct {
	Line  int    `json:"line"`
	Func  string `json:"function"`
	Short string `json:"short_description"`
	Long  string `json:"long_description"`
}

func RunChecker(r io.Reader) (Warnings, error) {
	cmd := exec.Command("check", checkOpts...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	io.Copy(stdin, r)
	stdin.Close()
	var warns Warnings
	if err := json.NewDecoder(stdout).Decode(&warns); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return warns, nil
}
