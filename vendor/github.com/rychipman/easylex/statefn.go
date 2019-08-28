package easylex

type StateFn func(*Lexer) StateFn
