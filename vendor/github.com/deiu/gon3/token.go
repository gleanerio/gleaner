package gon3

import (
	"github.com/rychipman/easylex"
)

const (
	// tokens expressed as literal strings in http://www.w3.org/TR/turtle/#sec-grammar-grammar
	tokenAtPrefix easylex.TokenType = iota
	tokenAtBase
	tokenSPARQLPrefix
	tokenSPARQLBase
	tokenEndTriples
	tokenA // 5
	tokenPredicateListSeparator
	tokenObjectListSeparator
	tokenStartBlankNodePropertyList
	tokenEndBlankNodePropertyList
	tokenStartCollection // 10
	tokenEndCollection
	tokenLiteralDatatypeTag
	tokenTrue
	tokenFalse

	// terminal tokens from http://www.w3.org/TR/turtle/#terminals
	tokenIRIRef // 15
	tokenPNameNS
	tokenPNameLN
	tokenBlankNodeLabel
	tokenLangTag
	tokenInteger // 20
	tokenDecimal
	tokenDouble
	tokenExponent
	tokenStringLiteralQuote
	tokenStringLiteralSingleQuote // 25
	tokenStringLiteralLongQuote
	tokenStringLiteralLongSingleQuote
	tokenAnon
)
