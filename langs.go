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
	return "unknown"
}

func (l *Lang) Set(s string) error {
	switch strings.ToLower(s) {
	case "c++", "cpp", "cc":
		*l = LangCpp
	case "go", "golang":
		*l = LangGo
	default:
		return errors.New("unknown language")
	}
	return nil
}

func (l *Lang) InlineCommentPrefix() string {
	switch *l {
	case LangCpp, LangGo:
		return " // "
	}
	return ""
}

func EncodeJsonAST(l Lang, r io.Reader, w io.Writer) error {
	switch l {
	case LangCpp:
		return encodeFromCpp(r, w)
	case LangGo:
		return encodeFromGo(r, w)
	default:
		return errors.New("unknown language")
	}
}
