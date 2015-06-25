/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"errors"
	"io"
	"strings"
)

type Lang int

const (
	LangAuto Lang = iota
	LangCpp
	LangGo
)

func (l *Lang) String() string {
	switch *l {
	case LangAuto:
		return "auto"
	case LangCpp:
		return "c++"
	case LangGo:
		return "go"
	}
	return ""
}

func (l *Lang) Set(s string) error {
	s = strings.ToLower(s)
	switch strings.ToLower(s) {
	case "c++", "cpp":
		*l = LangCpp
	case "go", "golang":
		*l = LangGo
	default:
		return errors.New("unknown language")
	}
	return nil
}

func EncodeJsonFromCode(l Lang, r io.Reader, w io.Writer) error {
	switch l {
	case LangCpp:
		return encodeJsonFromCppCode(r, w)
	case LangGo:
		return encodeJsonFromGoCode(r, w)
	default:
		return errors.New("unknown language")
	}
}
