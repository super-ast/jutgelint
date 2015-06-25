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
	LANG_CPP Lang = iota
	LANG_GO
)

func (l *Lang) String() string {
	switch *l {
	case LANG_CPP:
		return "c++"
	case LANG_GO:
		return "go"
	}
	return ""
}

func (l *Lang) Set(s string) error {
	s = strings.ToLower(s)
	switch strings.ToLower(s) {
	case "c++", "cpp":
		*l = LANG_CPP
	case "go", "golang":
		*l = LANG_GO
	default:
		return errors.New("unknown language")
	}
	return nil
}

func EncodeJsonFromCode(l Lang, r io.Reader, w io.Writer) error {
	switch l {
	case LANG_CPP:
		return encodeJsonFromCppCode(r, w)
	case LANG_GO:
		return encodeJsonFromGoCode(r, w)
	default:
		return errors.New("unknown language")
	}
}
