/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
)

var (
	checkOpts = []string{"/home/mvdan/output.json", "--dead-assign", "--fors", "--local-decl", "--variable-init"}

	lineRegex = regexp.MustCompile(`LINE: ([0-9]+)`)
)

type Result struct {
	Line  int
	Func  string
	Short string
	Long  string
}

func RunChecker(i io.Reader) ([]Result, error) {
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
	var results []Result
	var cur *Result
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		s := lineRegex.FindStringSubmatch(line)
		if s != nil {
			if cur != nil {
				results = append(results, *cur)
			}
			cur = &Result{}
			i, err := strconv.Atoi(s[1])
			if err != nil {
				return nil, err
			}
			cur.Line = i
		}
	}
	return results, nil
}
