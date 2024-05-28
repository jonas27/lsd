package cmd

import (
	"bytes"
	"fmt"
	"unicode"
)

func isJSON(s []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeftFunc(s, unicode.IsSpace), []byte{'{'})
}

func Lsd(in []byte) ([]byte, error) {
  fmt.Println(string(in))
  return nil, nil
}
