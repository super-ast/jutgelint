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
	lang = flag.String("lang", "", "Language to use (go, c++)")
)

func main() {
	flag.Parse()

	var json bytes.Buffer
	switch *lang {
	case "go":
		jutgelint.EncodeJsonFromGoCode(os.Stdin, &json)
	case "c++":
		log.Fatalf("unimplemented")
	default:
		log.Fatalf("unsupported language: '%s'", *lang)
	}
	results, err := jutgelint.RunChecker(&json)
	if err != nil {
		log.Fatalf("Error when running the checker: %v", err)
	}
	for _, r := range results {
		fmt.Printf("%#v\n", r)
	}
}
