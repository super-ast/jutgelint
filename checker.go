/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"
	"strconv"
)

var (
	checkOpts = []string{
		"/home/mvdan/output.json",
		"--dead-assign",
		"--fors",
		"--local-decl",
		"--variable-init",
	}

	lineRegex    = regexp.MustCompile(`LINE: ([0-9]+)`)
	funcRegex    = regexp.MustCompile(`FUNCTION: (.+)`)
	errDescRegex = regexp.MustCompile(`ERROR: ([^-]+) - (.+)`)
)

type Warning struct {
	Line  int
	Func  string
	Short string
	Long  string
}

func RunChecker(i io.Reader) ([]Warning, error) {
	//cmd := exec.Command("printer")
	cmd := exec.Command("check", checkOpts...)
	//stdin, err := cmd.StdinPipe()
	//if err != nil {
	//return nil, err
	//}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	//io.Copy(stdin, i)
	//stdin.Close()
	scanner := bufio.NewScanner(stdout)
	var warnings []Warning
	var cur *Warning
	for scanner.Scan() {
		line := scanner.Text()
		if s := lineRegex.FindStringSubmatch(line); s != nil {
			if cur != nil {
				warnings = append(warnings, *cur)
			}
			cur = &Warning{}
			i, err := strconv.Atoi(s[1])
			if err != nil {
				return nil, err
			}
			cur.Line = i
		} else if s := funcRegex.FindStringSubmatch(line); s != nil {
			cur.Func = s[1]
		} else if s := errDescRegex.FindStringSubmatch(line); s != nil {
			cur.Short = s[1]
			cur.Long = s[2]
		}

	}
	return warnings, nil
}
