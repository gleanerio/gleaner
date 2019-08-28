package jsonport

import "unsafe"

func ss(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func nn(b []byte) Number {
	return *(*Number)(unsafe.Pointer(&b))
}
