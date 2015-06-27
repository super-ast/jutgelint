/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mvdan/jutgelint"
)

var lang = jutgelint.LangAuto

func init() {
	flag.Var(&lang, "lang", "Language to use (auto, c++, go)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: jutgelint [input] [output]\n\n")
		fmt.Fprintf(os.Stderr, "The input and output files default to standard input and standard output\n")
		fmt.Fprintf(os.Stderr, "if none are specified.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "    jutgelint input.go output.go\n")
		fmt.Fprintf(os.Stderr, "    jutgelint -lang=cpp <input.cc >output.cc\n")
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) > 2 {
		flag.Usage()
		os.Exit(2)
	}

	in := os.Stdin
	out := os.Stdout

	if len(args) >= 1 {
		f, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("Cannot open file: %v", err)
		}
		if lang == jutgelint.LangAuto {
			var err error
			if lang, err = jutgelint.ParseLangFilename(args[0]); err != nil {
				log.Fatalf("Cannot infer language: %v", err)
			}
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

	if err := jutgelint.CheckAndCommentCode(lang, in, out); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
