package service

import (
	"io"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
)

func CharsetReader(set string, input io.Reader) (io.Reader, error) {
	return charset.NewReader("utf-8", input)
}
