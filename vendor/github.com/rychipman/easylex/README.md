# easylex

Easylex is a go library designed to simplify the process of lexing in go.
It was developed while working on a lexer and parser for the N3 file format, [gon3](http://github.com/rychipman/gon3).
Easylex aims to be simple, performant, and easily extensible.

## Design

Easylex borrows [Rob Pike's lexer design](https://cuddle.googlecode.com/hg/talk/lex.html) for go's native templates.
That design, however, quickly becomes cumbersome when dealing with a more complicated grammar; trying to lex a language with more complexity than go's relatively simple templates will quickly lead to repetitive, hard-to-read code.

