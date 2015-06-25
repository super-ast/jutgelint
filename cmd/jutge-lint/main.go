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

	code, err := ioutil.ReadAll(os.Stdin)
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
	jutgelint.CommentCode(warns, bytes.NewReader(code), os.Stdout)
}
