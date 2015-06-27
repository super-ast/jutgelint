/* Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	write = flag.Bool("write", false, "Write json results")
	name  = flag.String("name", "", "Test name")
)

func init() {
	flag.Parse()
}

const (
	testsDir = "tests"
	inGlob   = "in.*"
)

func getPathMatching(dir, file string) (string, error) {
	pattern := filepath.Join(testsDir, dir, file)
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(paths) != 1 {
		return "", errors.New("one matching file expected")
	}
	return paths[0], nil
}

func doTest(t *testing.T, name string) {
	inPath, err := getPathMatching(name, inGlob)
	if err != nil {
		t.Errorf("Failed opening in file: %v", err)
		return
	}
	base := filepath.Base(inPath)
	lang, err := ParseLangFilename(base)
	if err != nil {
		t.Errorf("Could not infer language: %v", err)
		return
	}
	outPath := filepath.Join(testsDir, name, "out"+filepath.Ext(base))
	in, err := os.Open(inPath)
	if err != nil {
		t.Errorf("Failed opening in file: %v", err)
		return
	}
	defer in.Close()
	var out bytes.Buffer
	if err := CheckAndCommentCode(lang, in, &out); err != nil {
		t.Errorf("Failed checking and commenting code: %v", err)
		return
	}
	got := out.Bytes()
	if *write {
		out, err := os.Create(outPath)
		if err != nil {
			t.Errorf("Failed opening out file: %v", err)
			return
		}
		defer out.Close()
		_, err = out.Write(got)
		if err != nil {
			t.Errorf("Failed writing out file: %v", err)
			return
		}
	} else {
		out, err := os.Open(outPath)
		if err != nil {
			t.Errorf("Failed opening out file: %v", err)
			return
		}
		defer out.Close()
		want, err := ioutil.ReadAll(out)
		if err != nil {
			t.Errorf("Failed reading out file: %v", err)
			return
		}
		if string(want) != string(got) {
			t.Errorf("Mismatching outputs in the test '%s'", name)
			return
		}
	}
}

func TestCases(t *testing.T) {
	entries, err := ioutil.ReadDir(testsDir)
	if err != nil {
		return
	}
	if *name != "" {
		doTest(t, *name)
	} else {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			doTest(t, e.Name())
		}
	}
}
