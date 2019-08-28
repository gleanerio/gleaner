package easylex

import (
	"fmt"
)

type TokenType int

const (
	TokenEOF   TokenType = -1
	TokenError TokenType = -2
)

type Token struct {
	Typ TokenType
	Val string
}

func (t Token) String() string {
	switch t.Typ {
	case TokenError:
		return t.Val
	case TokenEOF:
		return "EOF"
	}
	if len(t.Val) > 23 {
		return fmt.Sprintf("%.10q...%.10q", t.Val, t.Val[len(t.Val)-10:])
	}
	return fmt.Sprintf("%q", t.Val)
}
