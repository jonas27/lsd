package cmd

import (
	"bytes"
	"unicode"

	"github.com/jonas27/lsd/internal"
)

func isJSON(s []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeftFunc(s, unicode.IsSpace), []byte{'{'})
}

func Lsd(in []byte) (string, error) {
  return internal.Run(in)
}
