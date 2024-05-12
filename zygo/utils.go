package zygo

import (
	"strings"
	"unicode"
	"unsafe"
)

func SubStringLast(str, substring string) string {
	index := strings.Index(str, substring)
	if index == -1 {
		return ""
	}
	return str[index+len(substring):]
}
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)}))
}
