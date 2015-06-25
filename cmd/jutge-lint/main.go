/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/mvdan/jutgelint"
)

var (
	lang jutgelint.Lang
)

func init() {
	flag.Var(&lang, "lang", "Language to use (c++, go)")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) > 2 {
		flag.Usage()
	}

	in := os.Stdin
	out := os.Stdout

	if len(args) >= 1 {
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("Cannot open file: %v", err)
		}
		in = f
	}
	if len(args) >= 2 {
		f, err := os.Create(args[1])
		if err != nil {
			log.Fatalf("Cannot open file: %v", err)
		}
		out = f
	}

	code, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatalf("Error when reading code: %v", err)
	}
	var json bytes.Buffer
	if err := jutgelint.EncodeJsonFromCode(lang, bytes.NewReader(code), &json); err != nil {
		log.Fatalf("Could not translate code into json: %v", err)
	}

	warns, err := jutgelint.RunChecker(&json)
	if err != nil {
		log.Fatalf("Error when running the checker: %v", err)
	}
	jutgelint.CommentCode(warns, bytes.NewReader(code), out)
}
