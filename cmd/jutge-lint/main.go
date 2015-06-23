/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"bytes"
	"flag"
	"fmt"
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

	var json bytes.Buffer
	jutgelint.EncodeJsonFromCode(lang, os.Stdin, &json)
	results, err := jutgelint.RunChecker(&json)
	if err != nil {
		log.Fatalf("Error when running the checker: %v", err)
	}
	for _, r := range results {
		fmt.Printf("%#v\n", r)
	}
}
