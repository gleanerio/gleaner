package easylex

import (
	"fmt"
	"unicode/utf8"
)

const (
	EOF = rune(-1)
)

// Lexer is a struct that holds all private state
// necessare for lexing.
type Lexer struct {
	input  string
	state  StateFn
	start  int
	pos    int
	width  int
	tokens chan Token
}

// Lex returns a new Lexer instance that will lex the provided
// input starting from the provided state function.
func Lex(input string, state StateFn) *Lexer {
	return &Lexer{
		input:  input,
		state:  state,
		tokens: make(chan Token, 3), // TODO: troubleshoot buffer issues
	}
}

// NextToken returns the next token in the input
// currently being lexed.
func (l *Lexer) NextToken() Token {
	for {
		select {
		case tok := <-l.tokens:
			return tok
		default:
			if l.state == nil {
				break
			}
			l.state = l.state(l)
		}
	}
}

// Emit queues a token of the given type for retrieval by
// NextToken(). The token value is equal to all the runes
// processed since the last call to Emit() or Ignore().
func (l *Lexer) Emit(t TokenType) {
	l.tokens <- Token{
		t,
		l.input[l.start:l.pos],
	}
	l.start = l.pos
}

// Errorf emits an error token (a token of type TokenError)
// with a value equal to the formatted string.
func (l *Lexer) Errorf(format string, args ...interface{}) StateFn {
	l.tokens <- Token{
		TokenError,
		fmt.Sprintf(format, args),
	}
	return nil
}

// Next returns one rune and increments l.pos by the
// width of that rune. The width of the last rune
// processed is stored in l.width.
func (l *Lexer) Next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	var r rune
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// Backup decrements l.pos by the width of the last rune
// processed. Backup can only be called once per call to
// Next().
func (l *Lexer) Backup() {
	l.pos -= l.width
}

// Ignore resets l.start to the current value of l.pos.
// this ignores all the runes processed since the last
// call to Ignore() or Emit().
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// Peek returns the value of the rune at l.pos + 1, but
// does not mutate lexer state.
func (l *Lexer) Peek() rune {
	r := l.Next()
	l.Backup()
	return r
}
