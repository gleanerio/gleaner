package jsonport

import (
	"strconv"
)

type Type int

const (
	INVALID Type = iota
	OBJECT
	ARRAY
	STRING
	NUMBER
	BOOL
	NULL
)

func (t Type) String() string {
	switch t {
	case INVALID:
		return "INVALID"
	case OBJECT:
		return "OBJECT"
	case ARRAY:
		return "ARRAY"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case BOOL:
		return "BOOL"
	case NULL:
		return "NULL"
	}
	return "UNKNOWN"
}

// A Number represents a JSON number literal.
type Number string

// String returns the literal text of the number.
func (n Number) String() string { return string(n) }

// Float64 returns the number as a float64.
func (n Number) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

// Int64 returns the number as an int64.
func (n Number) Int64() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}
