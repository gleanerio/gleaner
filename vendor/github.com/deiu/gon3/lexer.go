package gon3

import (
	"github.com/rychipman/easylex"
)

type lexer interface {
	NextToken() easylex.Token
}

type mockLexer struct {
	tokens []easylex.Token
	pos    int
}

func newMockLexer(args ...easylex.Token) *mockLexer {
	return &mockLexer{
		tokens: args,
		pos:    0,
	}
}

func (m *mockLexer) NextToken() easylex.Token {
	ret := m.tokens[m.pos]
	m.pos += 1
	return ret
}

type typeMockLexer struct {
	types []easylex.TokenType
	pos   int
}

func newTypeMockLexer(args ...easylex.TokenType) *typeMockLexer {
	return &typeMockLexer{
		types: args,
		pos:   0,
	}
}

func (m *typeMockLexer) NextToken() easylex.Token {
	typ := m.types[m.pos]
	m.pos += 1
	return easylex.Token{
		typ,
		"n/a",
	}
}
