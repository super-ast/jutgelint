/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"encoding/json"
	"io"
	"os/exec"
)

const (
	CheckDeadAssign int = 1 << iota
	CheckFors
	CheckLocalDecl
	CheckVariableInit

	CheckAll int = -1
)

var optArgs = map[int]string{
	CheckDeadAssign:   "--dead-assign",
	CheckFors:         "--fors",
	CheckLocalDecl:    "--local-decl",
	CheckVariableInit: "--variable-init",
}

type Warnings map[string][]Warning

type Warning struct {
	Line  int    `json:"line"`
	Func  string `json:"function"`
	Short string `json:"short_description"`
	Long  string `json:"long_description"`
}

func getCheckOpts(checks int) []string {
	var opts []string
	for c := range optArgs {
		if checks&c > 0 {
			opts = append(opts, optArgs[c])
		}
	}
	return opts
}

func RunChecker(r io.Reader, checks int) (Warnings, error) {
	cmd := exec.Command("check", getCheckOpts(checks)...)
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
	if _, err := io.Copy(stdin, r); err != nil {
		return nil, err
	}
	if err := stdin.Close(); err != nil {
		return nil, err
	}
	var warns Warnings
	if err := json.NewDecoder(stdout).Decode(&warns); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return warns, nil
}
