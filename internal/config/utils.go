package config

import (
	"path/filepath"
	"strings"
)

func fileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// moveToFront moves needle to the front of haystack, in place if possible.
// from https://github.com/golang/go/wiki/SliceTricks
func MoveToFront(needle string, haystack []string) []string {
	if len(haystack) != 0 && haystack[0] == needle {
		return haystack
	}
	prev := needle
	for i, elem := range haystack {
		switch {
		case i == 0:
			haystack[0] = needle
			prev = elem
		case elem == needle:
			haystack[i] = prev
			return haystack
		default:
			haystack[i] = prev
			prev = elem
		}
	}
	return append(haystack, prev)
}
