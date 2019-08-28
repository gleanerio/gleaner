package jsonport

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	jsontrue  = []byte("true")
	jsonfalse = []byte("false")
	jsonnull  = []byte("null")
)

func isspace(b byte) bool {
	switch b {
	case ' ':
	case '\t':
	case '\n':
	case '\r':
	case '\v':
	case '\f':
	default:
		return false
	}
	return true
}

func skipspace(b []byte) int {
	for i, c := range b {
		if !isspace(c) {
			return i
		}
	}
	return len(b)
}

func parse(b []byte) (Json, int, error) {
	i := skipspace(b)
	b = b[i:]
	if len(b) == 0 {
		return Json{}, 0, errJSONEOF
	}

	var j Json
	switch b[0] {
	case '{':
		o, ii, err := parseObject(b, false)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.m = o
		j.tp = OBJECT
	case '[':
		a, ii, err := parseArray(b)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.a = a
		j.tp = ARRAY
	case '"':
		s, ii, err := parseString(b)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.b = s
		j.tp = STRING
	case 't', 'f':
		tf, ii, err := parseBool(b)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.t = tf
		j.tp = BOOL
	case 'n':
		ii, err := parseNull(b)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.tp = NULL
	default:
		n, ii, err := parseNumber(b)
		if err != nil {
			j.err = err
			return j, i, err
		}
		i += ii
		j.b = n
		j.tp = NUMBER
	}
	return j, i, nil
}

func parseMemberName(k interface{}) (string, error) {
	switch t := k.(type) {
	case string:
		return t, nil
	}
	return "", errMemberName
}

func parseArrayIndex(k interface{}) (int, error) {
	switch t := k.(type) {
	// reflect.ValueOf(t).Int() or Uint() ?
	// without reflection here.
	case int:
		return t, nil
	case int8:
		return int(t), nil
	case int16:
		return int(t), nil
	case int32:
		return int(t), nil
	case int64:
		return int(t), nil
	case uint:
		return int(t), nil
	case uint8:
		return int(t), nil
	case uint16:
		return int(t), nil
	case uint32:
		return int(t), nil
	case uint64:
		return int(t), nil
	}
	return 0, errArrayIndex

}

func parsePath(b []byte, keys ...interface{}) (Json, int, error) {
	if len(keys) == 0 {
		return parse(b)
	}

	i := skipspace(b)
	b = b[i:]
	if len(b) == 0 {
		return Json{}, i, errJSONEOF
	}

	if name, err := parseMemberName(keys[0]); err == nil {
		if name == ParseMemberNamesOnly {
			o, ii, err := parseObject(b, true)
			i += ii
			j := Json{m: o, tp: OBJECT}
			return j, i, err
		}
		j, ii, err := parseObjectMember(b, name, keys[1:]...)
		i += ii
		return j, i, err
	}

	if index, err := parseArrayIndex(keys[0]); err == nil {
		j, ii, err := parseArrayElement(b, index, keys[1:]...)
		i += ii
		return j, i, err
	} else {
		return Json{}, 0, errors.New("key type error")
	}
}

func parseString(b []byte) ([]byte, int, error) {
	if b[0] != '"' {
		return nil, 0, fmt.Errorf("STRING: expect '\"' found '%c'", b[0])
	}
	var i int
	var escaped bool
	for i, c := range b[1:] {
		if c == '\\' {
			escaped = !escaped
		} else if c == '"' && !escaped {
			s := b[1 : i+1] // trim "\""
			return s, i + 2, nil
		} else {
			escaped = false
		}
	}
	return nil, i, errStringEOF
}

type obj struct {
	kvs []kv
}

var opool = sync.Pool{
	New: func() interface{} {
		return &obj{kvs: make([]kv, 0, 1000)}
	},
}

func parseObject(b []byte, namesonly bool) ([]kv, int, error) {
	if len(b) == 0 {
		return nil, 0, errors.New("OBJECT: expect '{' found EOF")
	}
	if b[0] != '{' {
		return nil, 0, fmt.Errorf("OBJECT: expect '{' found '%c", b[0])
	}
	if len(b) < 2 {
		return nil, 1, errors.New("OBJECT: expect '}' found EOF")
	}
	if b[1] == '}' {
		return nil, 2, nil
	}

	const (
		stateMemberName  = 1
		stateColon       = 2
		stateMemberValue = 3
		stateDone        = 4
	)
	state := stateMemberName

	var k []byte

	p := opool.Get().(*obj)
	defer opool.Put(p)
	p.kvs = p.kvs[:0]

	i := 1 // skip {
	for i < len(b) {
		if isspace(b[i]) {
			i++
			continue
		}
		if state == stateMemberName {
			b, ii, err := parseString(b[i:])
			if err != nil {
				return nil, i, fmt.Errorf("OBJECT member.name: %s", err)
			}
			i += ii
			k = b
			state = stateColon
			continue
		}
		if state == stateColon {
			if b[i] != ':' {
				return nil, i, fmt.Errorf("OBJECT: expect ':' found '%c'", b[i])
			}
			i++
			state = stateMemberValue
			continue
		}
		if state == stateMemberValue {
			j := Json{tp: NULL}
			var ii int
			var err error
			if namesonly {
				ii, err = jsonskip(b[i:])
			} else {
				j, ii, err = parse(b[i:])
				if err != nil {
					return nil, i, fmt.Errorf("OBJECT: member %q parse err: %s", k, err)
				}
			}
			i += ii
			p.kvs = append(p.kvs, kv{k: k, v: j})
			state = stateDone
			continue
		}
		if state == stateDone {
			if b[i] == ',' {
				i++
				state = stateMemberName
				continue
			}
			if b[i] == '}' {
				i++
				m := make([]kv, 0, len(p.kvs))
				return append(m, p.kvs...), i, nil
			}
			return nil, i, fmt.Errorf("OBJECT: expect ',' or '}' found '%c'", b[i])
		}
	}
	return nil, i, errors.New("OBJECT: internal err")
}

func parseArray(b []byte) ([]Json, int, error) {
	if len(b) == 0 {
		return nil, 0, errors.New("ARRAY: expect '[' found EOF")
	}
	if b[0] != '[' {
		return nil, 0, fmt.Errorf("ARRAY: expect '[' found '%c'", b[0])
	}
	if len(b) < 2 {
		return nil, 1, errors.New("ARRAY: expect ']' found EOF")
	}
	if b[1] == ']' {
		return []Json{}, 2, nil
	}

	const (
		stateValue = 1
		stateDone  = 2
	)
	state := stateValue

	var a []Json

	i := 1 // skip [
	for i < len(b) {
		if isspace(b[i]) {
			i++
			continue
		}
		if state == stateValue {
			j, ii, err := parse(b[i:])
			if err != nil {
				return a, i, fmt.Errorf("ARRAY: index %d value: %s", len(a), err)
			}
			i += ii
			a = append(a, j)
			state = stateDone
			continue
		}

		if state == stateDone {
			if b[i] == ',' {
				i++
				state = stateValue
				continue
			}
			if b[i] == ']' {
				i++
				return a, i, nil
			}
			return nil, i, fmt.Errorf("ARRAY: expect ',' or ']' found '%c'", b[i])
		}
	}
	return nil, i, errArrayEOF
}

func parseNumber(b []byte) ([]byte, int, error) {
	if len(b) == 0 {
		return nil, 0, errJSONEOF
	}
	c := b[0]
	if c != '-' && (c < '0' || c > '9') {
		return nil, 0, errors.New("Unknown type")
	}
	var i int
	for ; i < len(b); i++ {
		c := b[i]
		switch {
		case c >= '0' && c <= '9':
		case c == '.':
		case c == 'e':
		case c == 'E':
		case c == '+':
		case c == '-':
		default:
			return b[:i], i, nil
		}
	}
	return b[:i], i, nil
}

func parseBool(b []byte) (bool, int, error) {
	if bytes.HasPrefix(b, jsontrue) {
		return true, len(jsontrue), nil
	}
	if bytes.HasPrefix(b, jsonfalse) {
		return false, len(jsonfalse), nil
	}
	return false, 0, errors.New("BOOL: not true nor false")
}

func parseNull(b []byte) (int, error) {
	if bytes.HasPrefix(b, jsonnull) {
		return len(jsonnull), nil
	}
	return 0, errors.New("NULL: parse err")
}

func parseObjectMember(b []byte, name string, keys ...interface{}) (Json, int, error) {
	if len(b) == 0 {
		return Json{}, 0, errors.New("OBJECT: expect '{' found EOF")
	}
	if b[0] != '{' {
		return Json{}, 0, fmt.Errorf("OBJECT: expect '{' found '%c", b[0])
	}
	if len(b) < 2 {
		return Json{}, 1, errors.New("OBJECT: expect '}' found EOF")
	}
	if b[1] == '}' {
		return Json{tp: NULL}, 2, nil
	}

	const (
		stateMemberName  = 1
		stateColon       = 2
		stateMemberValue = 3
		stateDone        = 4
	)
	state := stateMemberName

	var k string

	i := 1 // skip {
	for i < len(b) {
		if isspace(b[i]) {
			i++
			continue
		}
		if state == stateMemberName {
			s, ii, err := parseString(b[i:])
			if err != nil {
				return Json{}, i, fmt.Errorf("OBJECT member.name: %s", err)
			}
			i += ii
			k = unquote(s)
			state = stateColon
			continue
		}
		if state == stateColon {
			if b[i] != ':' {
				return Json{}, i, fmt.Errorf("OBJECT: expect ':' found '%c'", b[i])
			}
			i++
			state = stateMemberValue
			continue
		}

		if state == stateMemberValue {
			if k == name {
				j, ii, err := parsePath(b[i:], keys...)
				if err != nil {
					return j, i, fmt.Errorf("OBJECT member %q parse err: %s", k, err)
				}
				return j, i + ii, nil
			} else {
				ii, err := jsonskip(b[i:])
				if err != nil {
					return Json{}, i, fmt.Errorf("OBJECT member: %q parse err: %s", k, err)
				}
				i += ii
				state = stateDone
			}
			continue
		}

		if state == stateDone {
			if b[i] == ',' {
				i++
				state = stateMemberName
				continue
			}
			if b[i] == '}' {
				i++
				return Json{tp: NULL}, i, nil
			}
			return Json{}, i, fmt.Errorf("OBJECT: expect ',' or '}' found '%c'", b[i])
		}
	}
	return Json{}, i, errObjectEOF
}

func parseArrayElement(b []byte, index int, keys ...interface{}) (Json, int, error) {
	if len(b) == 0 {
		return Json{}, 0, errors.New("ARRAY: expect '[' found EOF")
	}
	if b[0] != '[' {
		return Json{}, 0, fmt.Errorf("ARRAY: expect '[' found '%c'", b[0])
	}
	if len(b) < 2 {
		return Json{}, 1, errors.New("ARRAY: expect ']' found EOF")
	}
	if b[1] == ']' {
		return Json{}, 2, nil
	}

	if index < 0 {
		return Json{tp: NULL}, 0, nil
	}

	const (
		stateValue = 1
		stateDone  = 2
	)
	state := stateValue

	pos := 0

	i := 1 // skip [
	for i < len(b) {
		if isspace(b[i]) {
			i++
			continue
		}
		if state == stateValue {
			if pos != index {
				ii, err := jsonskip(b[i:])
				if err != nil {
					return Json{}, i, fmt.Errorf("ARRAY: index %d err: %s", pos, err)
				}
				i += ii
			} else {
				j, ii, err := parsePath(b[i:], keys...)
				if err != nil {
					return Json{}, i, fmt.Errorf("ARRAY: index %d err: %s", pos, err)
				}
				return j, i + ii, nil
			}
			pos += 1
			state = stateDone
			continue
		}

		if state == stateDone {
			if b[i] == ',' {
				i++
				state = stateValue
				continue
			}
			if b[i] == ']' {
				i++
				return Json{tp: NULL}, i, nil
			}
			return Json{}, i, fmt.Errorf("ARRAY: expect ',' or ']' found '%c'", b[i])
		}
	}
	return Json{}, i, errArrayEOF
}

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func unquote(s []byte) string {
	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return ss(s)
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room?  Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return ""
			}
			switch s[r] {
			default:
				return ""
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return ""
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return ""

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return string(b[0:w])
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(ss(s[2:6]), 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}
