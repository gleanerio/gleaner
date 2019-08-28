package gon3

import (
	"github.com/rychipman/easylex"
	"strings"
)

const (
	eof = -1
)

func lexDocument(l *easylex.Lexer) easylex.StateFn {
	matchWhitespace.MatchRun(l)
	l.Ignore()
	switch l.Peek() {
	case easylex.EOF:
		l.Emit(easylex.TokenEOF)
		return nil
	case '#':
		return lexComment
	case '@':
		return lexAtStatement
	case '_':
		return lexBlankNodeLabel
	case '<':
		return lexIRIRef
	case '"', '\'':
		return lexRDFLiteral
	case '.':
		if matchBareDecimalStart.Peek(l) {
			return lexNumericLiteral
		}
		return lexPunctuation
	case '+', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return lexNumericLiteral
	case '[':
		// anon
		l.Next()
		matchWhitespace.MatchRun(l)
		if matchCloseBracket.MatchOne(l) {
			l.Emit(tokenAnon)
			return lexDocument
		}
		l.Emit(tokenStartBlankNodePropertyList)
		return lexDocument
	case '^', ']', '(', ')', ';', ',':
		return lexPunctuation
	case 't', 'f', 'a', 'b', 'B', 'p', 'P':
		if matchTrue.MatchOne(l) {
			if isWhitespace(l.Peek()) {
				l.Emit(tokenTrue)
				return lexDocument
			}
		}
		if matchFalse.MatchOne(l) {
			if isWhitespace(l.Peek()) {
				l.Emit(tokenFalse)
				return lexDocument
			}
		}
		if matchA.MatchOne(l) {
			if isWhitespace(l.Peek()) {
				l.Emit(tokenA)
				return lexDocument
			}
		}
		if matchSPARQLBase.MatchOne(l) {
			if isWhitespace(l.Peek()) {
				l.Emit(tokenSPARQLBase)
				return lexDocument
			}
		}
		if matchSPARQLPrefix.MatchOne(l) {
			if isWhitespace(l.Peek()) {
				l.Emit(tokenSPARQLPrefix)
				return lexDocument
			}
		}
		fallthrough
	default:
		return lexPName
	}
	panic("unreachable")
}

func isWhitespace(r rune) bool {
	if strings.IndexRune("\n\r\t\v\f ", r) >= 0 {
		return true
	}
	return false
}

func lexComment(l *easylex.Lexer) easylex.StateFn {
	easylex.NewMatcher().AcceptRunes("#").AssertOne(l, "Expected '#' while lexing comment")
	easylex.NewMatcher().RejectRunes("\n").MatchRun(l)
	l.Ignore()
	return lexDocument
}

func lexAtStatement(l *easylex.Lexer) easylex.StateFn {
	easylex.NewMatcher().AcceptRunes("@").AssertOne(l, "Expected '@' while lexing AtStatement")
	if easylex.NewMatcher().AcceptString("prefix").MatchOne(l) {
		if isWhitespace(l.Peek()) {
			l.Emit(tokenAtPrefix)
			return lexDocument
		}
	}
	if easylex.NewMatcher().AcceptString("base").MatchOne(l) {
		if isWhitespace(l.Peek()) {
			l.Emit(tokenAtBase)
			return lexDocument
		}
	}
	matchAlphabet.AssertRun(l, "Expected alphabet while lexing AtStatement")
	for {
		hyphen := easylex.NewMatcher().AcceptRunes("-").MatchOne(l)
		alph := matchAlphaNumeric.MatchRun(l)
		if !hyphen && !alph {
			break
		}
		if hyphen != alph {
			// TODO: error
		}
	}
	l.Emit(tokenLangTag)
	return lexDocument
}

func lexBlankNodeLabel(l *easylex.Lexer) easylex.StateFn {
	easylex.NewMatcher().AcceptRunes("_").AssertOne(l, "Expected '_' while lexing bnode label")
	easylex.NewMatcher().AcceptRunes(":").AssertOne(l, "Expected ':' while lexing bnode label")
	easylex.NewMatcher().Union(matchPNCharsU).Union(matchDigits).AssertOne(l, "Expected pncharsu or digit while lexing bnode label")
	easylex.NewMatcher().Union(matchPNChars).AcceptRunes(".").MatchLookAheadRun(l, easylex.NewMatcher().Union(matchPNChars).AcceptRunes("."))
	matchPNChars.MatchRun(l)
	l.Emit(tokenBlankNodeLabel)
	return lexDocument
}

func lexIRIRef(l *easylex.Lexer) easylex.StateFn {
	easylex.NewMatcher().AcceptRunes("<").AssertOne(l, "Expected '<' while lexing iriref")
	iriChars := easylex.NewMatcher().RejectRunes("<>\"{}|^`\\\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u0008\u0009\u000a\u000b\u000c\u000d\u000e\u000f\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001a\u001b\u001c\u001d\u001e\u001f\u0020")
	for {
		m1 := iriChars.MatchRun(l)
		if l.Peek() == '\\' {
			l.Next()
			if l.Peek() == 'u' {
				l.Next()
				for i := 0; i < 4; i += 1 {
					matchHex.AssertOne(l, "Expected hex digit while lexing iriref")
				}
			} else if l.Peek() == 'U' {
				l.Next()
				for i := 0; i < 8; i += 1 {
					matchHex.AssertOne(l, "Expected hex digit while lexing iriref")
				}
			} else {
				// TODO: error
			}
		} else if !m1 {
			break
		}
	}
	easylex.NewMatcher().AcceptRunes(">").AssertOne(l, "Expected '>' while lexing iriref")
	l.Emit(tokenIRIRef)
	return lexDocument
}

func lexRDFLiteral(l *easylex.Lexer) easylex.StateFn {
	if matchLongQuote.MatchOne(l) {
		for {
			if matchLongQuote.MatchOne(l) {
				break
			}
			q := matchQuote.MatchOne(l)
			q = matchQuote.MatchOne(l) || q
			ch := true
			if easylex.NewMatcher().RejectRunes(`"\`).MatchOne(l) {
				// do nothing
			} else if l.Peek() == '\\' {
				l.Next()
				switch l.Peek() {
				case 'u':
					l.Next()
					for i := 0; i < 4; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 'U':
					l.Next()
					for i := 0; i < 8; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 't', 'b', 'n', 'r', 'f', '"', '\'', '\\':
					l.Next()
				default:
					ch = false
				}
			}
			if q && !ch {
				// TODO: error
			}
		}
		l.Emit(tokenStringLiteralLongQuote)
		return lexDocument
	}
	if matchLongSingleQuote.MatchOne(l) {
		for {
			if matchLongSingleQuote.MatchOne(l) {
				break
			}
			q := matchSingleQuote.MatchOne(l)
			q = matchSingleQuote.MatchOne(l) || q
			ch := true
			if easylex.NewMatcher().RejectRunes(`'\`).MatchOne(l) {
				// do nothing
			} else if l.Peek() == '\\' {
				l.Next()
				switch l.Peek() {
				case 'u':
					l.Next()
					for i := 0; i < 4; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 'U':
					l.Next()
					for i := 0; i < 8; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 't', 'b', 'n', 'r', 'f', '"', '\'', '\\':
					l.Next()
				default:
					ch = false
				}
			}
			if !q && !ch {
				break
			}
		}
		l.Emit(tokenStringLiteralLongSingleQuote)
		return lexDocument
	}
	if matchQuote.MatchOne(l) {
		for {
			if easylex.NewMatcher().RejectRunes("\u0022\u005c\u000a\u000d").MatchOne(l) {
				// do nothing
			} else if l.Peek() == '\\' {
				l.Next()
				switch l.Peek() {
				case 'u':
					l.Next()
					for i := 0; i < 4; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 'U':
					l.Next()
					for i := 0; i < 8; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 't', 'b', 'n', 'r', 'f', '"', '\'', '\\':
					l.Next()
				default:
					break
				}
			} else {
				break
			}
		}
		matchQuote.AssertOne(l, "Expected '\"' while lexing RDF Literal")
		l.Emit(tokenStringLiteralQuote)
		return lexDocument
	}
	if matchSingleQuote.MatchOne(l) {
		for {
			if easylex.NewMatcher().RejectRunes("\u0027\u005c\u000a\u000d").MatchOne(l) {
				// do nothing
			} else if l.Peek() == '\\' {
				l.Next()
				switch l.Peek() {
				case 'u':
					l.Next()
					for i := 0; i < 4; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 'U':
					l.Next()
					for i := 0; i < 8; i += 1 {
						matchHex.AssertOne(l, "Expected hex digit while lexing RDF Literal")
					}
				case 't', 'b', 'n', 'r', 'f', '"', '\'', '\\':
					l.Next()
				default:
					break
				}
			} else {
				break
			}
		}
		matchSingleQuote.AssertOne(l, "Expected \"'\" while lexing RDF Literal")
		l.Emit(tokenStringLiteralSingleQuote)
		return lexDocument
	}
	panic("unreachable")
}

func lexNumericLiteral(l *easylex.Lexer) easylex.StateFn {
	easylex.NewMatcher().AcceptRunes("+-").MatchOne(l)
	if matchDigits.MatchRun(l) {
		if easylex.NewMatcher().AcceptRunes("eE").MatchOne(l) {
			easylex.NewMatcher().AcceptRunes("+-").MatchOne(l)
			matchDigits.AssertRun(l, "Expected digits while lexing numeric literal")
			l.Emit(tokenDouble)
			return lexDocument
		} else if matchPeriod.MatchLookAhead(l, easylex.NewMatcher().AcceptRunes("0123456789eE")) {
			if matchDigits.MatchRun(l) {
				if isWhitespace(l.Peek()) {
					l.Emit(tokenDecimal)
					return lexDocument
				}
			}
			easylex.NewMatcher().AcceptRunes("eE").AssertOne(l, "Expected 'e' or 'E' while lexing numeric literal")
			easylex.NewMatcher().AcceptRunes("+-").MatchOne(l)
			matchDigits.AssertRun(l, "Expected digits while lexing numeric literal")
			l.Emit(tokenDouble)
			return lexDocument
		} else {
			l.Emit(tokenInteger)
			return lexDocument
		}
	} else {
		matchPeriod.AssertOne(l, "Expected '.' while lexing numeric literal")
		matchDigits.AssertRun(l, "Expected digits while lexing numeric literal")
		if easylex.NewMatcher().AcceptRunes("eE").MatchOne(l) {
			easylex.NewMatcher().AcceptRunes("+-").MatchOne(l)
			matchDigits.AssertRun(l, "Expected digits while lexing numeric literal")
			l.Emit(tokenDouble)
			return lexDocument
		}
		l.Emit(tokenDecimal)
		return lexDocument
	}
}

func lexPunctuation(l *easylex.Lexer) easylex.StateFn {
	// ^ ] ( ) ; , .
	if matchDoubleCaret.MatchOne(l) {
		l.Emit(tokenLiteralDatatypeTag)
	} else if matchCloseBracket.MatchOne(l) {
		l.Emit(tokenEndBlankNodePropertyList)
	} else if matchOpenParens.MatchOne(l) {
		l.Emit(tokenStartCollection)
	} else if matchCloseParens.MatchOne(l) {
		l.Emit(tokenEndCollection)
	} else if matchSemicolon.MatchOne(l) {
		l.Emit(tokenPredicateListSeparator)
	} else if matchComma.MatchOne(l) {
		l.Emit(tokenObjectListSeparator)
	} else if matchPeriod.MatchOne(l) {
		l.Emit(tokenEndTriples)
	} else {
		// TODO: error
	}
	return lexDocument
}

func lexPName(l *easylex.Lexer) easylex.StateFn {
	// accept PN_PREFIX
	matchPNCharsBase.MatchOne(l)
	for {
		period := matchPeriod.MatchRun(l)
		pnchars := matchPNChars.MatchRun(l)
		if !pnchars {
			if period {
				// TODO: error
			}
			break
		}
	}

	easylex.NewMatcher().AcceptRunes(":").AssertOne(l, "Expected ':' while lexing pname")
	// TODO: get exhaustive list of "end" chars
	// TODO: factor this out into a matcher
	if easylex.NewMatcher().AcceptRunes("\n\r\t\v\f ;,.#").Peek(l) {
		l.Emit(tokenPNameNS)
		return lexDocument
	}
	// accept PN_LOCAL
	if l.Peek() == '\\' {
		l.Next()
		matchEscapable.AssertOne(l, "Expected escapable while lexing pname")
	} else if l.Peek() == '%' {
		l.Next()
		matchHex.AssertOne(l, "Expected hex digit while lexing pname")
		matchHex.AssertOne(l, "Expected hex digit while lexing pname")
	} else {
		easylex.NewMatcher().AcceptRunes(":").Union(matchPNCharsU).Union(matchDigits).AssertOne(l, "Expected ':', pncharsu, or digits while lexing pname")
	}
	for {
		m := easylex.NewMatcher().Union(matchPNChars).AcceptRunes(".:").MatchLookAheadRun(l, easylex.NewMatcher().Union(matchPNChars).AcceptRunes(`.:\%`))
		for {
			switch l.Peek() {
			case '\\':
				l.Next()
				matchEscapable.AssertOne(l, "Expected escapable while lexing pname")
				continue
			case '%':
				l.Next()
				matchHex.AssertOne(l, "Expected hex digit while lexing pname")
				matchHex.AssertOne(l, "Expected hex digit while lexing pname")
				continue
			}
			break
		}
		if !m {
			break
		}
	}
	easylex.NewMatcher().Union(matchPNChars).AcceptRunes(":").MatchRun(l)
	l.Emit(tokenPNameLN)
	return lexDocument
}
