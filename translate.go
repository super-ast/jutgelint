/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package jutgelint

import (
	"errors"
	"io"
)

type Lang int

const (
	LANG_CPP Lang = iota
	LANG_GO
)

func EncodeJsonFromCode(l Lang, i io.Reader, w io.Writer) error {
	switch l {
	case LANG_CPP:
		return encodeJsonFromCppCode(i, w)
	case LANG_GO:
		return encodeJsonFromGoCode(i, w)
	default:
		return errors.New("unknown language")
	}
}
