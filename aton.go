/*
at object notation

GoFmt GoBuildNull
GoRun
*/

package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	NL = "\n"
)

type Value interface{}

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenKey
	tokenString
	tokenInteger
	tokenDictOpen
	tokenDictClose
	tokenListOpen
	tokenListClose
)

type token struct {
	typ   tokenType
	value string
}

type lexer struct {
	input string
	pos   int
	ch    rune
}

func newLexer(input string) *lexer {
	l := &lexer{input: input}
	l.readChar()
	return l
}

func (l *lexer) readChar() {
	if l.pos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.pos])
	}
	l.pos++
}

func (l *lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *lexer) readBracketedString() string {
	var result strings.Builder
	l.readChar() // skip '['

	for l.ch != ']' && l.ch != 0 {
		result.WriteRune(l.ch)
		l.readChar()
	}

	if l.ch == ']' {
		l.readChar() // skip ']'
	}

	return result.String()
}

func (l *lexer) readInteger() string {
	var result strings.Builder
	l.readChar() // skip '<'

	for unicode.IsDigit(l.ch) {
		result.WriteRune(l.ch)
		l.readChar()
	}

	if l.ch == '>' {
		l.readChar() // skip '>'
	}

	return result.String()
}

func (l *lexer) readUnquotedString() string {
	var result strings.Builder

	for l.ch != 0 && !unicode.IsSpace(l.ch) &&
		l.ch != '{' && l.ch != '}' &&
		l.ch != '(' && l.ch != ')' &&
		l.ch != '<' && l.ch != '[' {
		result.WriteRune(l.ch)
		l.readChar()
	}

	return result.String()
}

func (l *lexer) nextToken() token {
	l.skipWhitespace()

	switch l.ch {
	case 0:
		return token{typ: tokenEOF}
	case '{':
		l.readChar()
		return token{typ: tokenDictOpen}
	case '}':
		l.readChar()
		return token{typ: tokenDictClose}
	case '(':
		l.readChar()
		return token{typ: tokenListOpen}
	case ')':
		l.readChar()
		return token{typ: tokenListClose}
	case '@':
		l.readChar()
		key := l.readUnquotedString()
		return token{typ: tokenKey, value: key}
	case '[':
		str := l.readBracketedString()
		return token{typ: tokenString, value: str}
	case '<':
		intStr := l.readInteger()
		return token{typ: tokenInteger, value: intStr}
	default:
		// Unquoted string
		str := l.readUnquotedString()
		return token{typ: tokenString, value: str}
	}
}

type Parser struct {
	lexer     *lexer
	curToken  token
	peekToken token
}

func NewParser(input string) *Parser {
	p := &Parser{lexer: newLexer(input)}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.nextToken()
}

func (p *Parser) Parse() (map[string]interface{}, error) {
	return p.parseDict()
}

func (p *Parser) parseDict() (map[string]interface{}, error) {
	dict := make(map[string]interface{})

	// Skip opening brace if present
	if p.curToken.typ == tokenDictOpen {
		p.nextToken()
	}

	for p.curToken.typ != tokenDictClose && p.curToken.typ != tokenEOF {
		if p.curToken.typ != tokenKey {
			return nil, fmt.Errorf("expected key, got %v", p.curToken)
		}

		key := p.curToken.value
		p.nextToken()

		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		dict[key] = value
	}

	if p.curToken.typ == tokenDictClose {
		p.nextToken()
	}

	return dict, nil
}

func (p *Parser) parseList() ([]interface{}, error) {
	list := make([]interface{}, 0)

	if p.curToken.typ == tokenListOpen {
		p.nextToken()
	}

	for p.curToken.typ != tokenListClose && p.curToken.typ != tokenEOF {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		list = append(list, value)
	}

	if p.curToken.typ == tokenListClose {
		p.nextToken()
	}

	return list, nil
}

func (p *Parser) parseValue() (Value, error) {
	switch p.curToken.typ {
	case tokenDictOpen:
		return p.parseDict()
	case tokenListOpen:
		return p.parseList()
	case tokenInteger:
		val, err := strconv.Atoi(p.curToken.value)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %s", p.curToken.value)
		}
		p.nextToken()
		return val, nil
	case tokenString:
		val := p.curToken.value
		p.nextToken()
		return val, nil
	default:
		return nil, fmt.Errorf("unexpected token %v", p.curToken)
	}
}

func ParseDocument(input string) (map[string]interface{}, error) {
	parser := NewParser(input)
	return parser.Parse()
}
func main() {

	input := `
    @version <2>
    @models (
        {
            @name my_transformation
            @description [This model transforms raw data]
            @columns (
                { @name id @description [A unique identifier] }
                { @name name }
            )
        }
    )`

	doc, err := ParseDocument(input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v"+NL, doc)

}
