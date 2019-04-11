package parser

import (
	"bufio"
	"io"
)

// SyntacticParser performs lexical parsing of the input according definition
// of some lexemes like terminal or non-terminal symbol.
type SyntacticParser struct {
	Reader io.Reader

	buf []byte
	pos int
}

func NewSyntacticParser(reader io.Reader) *SyntacticParser {
	return &SyntacticParser{Reader: reader}
}

func (p *SyntacticParser) Parse() (*AST, error) {
	if lemmes, err := p.parseSyntax(); err != nil {
		return nil, &Error{err, p.pos + 1}
	} else {
		return &AST{lemmes: lemmes, semantic: false}, nil
	}
}

func (p *SyntacticParser) eof() error {
	if p.pos == len(p.buf) {
		return io.EOF
	} else {
		return nil
	}
}

func (p *SyntacticParser) parseSyntax() ([][]Node, error) {
	var rules [][]Node
	var scanner = bufio.NewScanner(p.Reader)

	for scanner.Scan() {
		// Reset parser state with the new line.
		p.buf = []byte(scanner.Text())
		p.pos = 0

		// Parse every single line and ignore parsing errors.
		if rule, err := p.parseRule(); err == nil {
			rules = append(rules, rule)
		}
	}

	return rules, scanner.Err()
}

func (p *SyntacticParser) parseComment() (*Comment, error) {
	if err := p.eof(); err != nil {
		return nil, err
	}

	var token = Token{Begin: p.pos}

	if p.buf[p.pos] != ';' {
		return nil, ErrUnexpectedChar
	}

	for token.End = p.pos + 1; token.End != len(p.buf); token.End++ {
		if p.buf[token.End] == '\n' || p.buf[token.End] == byte(0) {
			break
		}
	}

	p.pos = token.End
	return &Comment{token}, nil
}

func (p *SyntacticParser) parseRule() ([]Node, error) {
	var tokens []Node

	// Try to apply lexeme parser to every position at line. Since parsing of
	// simplified grammar which contains only operators `::=` and `|` and
	// terminal and non-terminal symbols requires only one look-a-head
	// character, then we can skip parsing other tokens if the current parsing
	// attempt were successfull.
	for p.pos < len(p.buf) {
		if tok, err := p.parseDisjunction(); err == nil {
			var expr = Expression{Token: *tok}
			tokens = append(tokens, &AlternativeExpression{expr})
			continue
		}

		if tok, err := p.parseDefinitionSimbol(); err == nil {
			var expr = Expression{Token: *tok}
			tokens = append(tokens, &AssignmentExpression{expr})
			continue
		}

		if tok, err := p.parseAtom(); err == nil {
			tokens = append(tokens, tok)
			continue
		}

		if tok, err := p.parseComment(); err == nil {
			tokens = append(tokens, tok)
			continue
		}

		p.pos++
	}

	return tokens, nil
}

func (p *SyntacticParser) parseRuleName() ([]byte, error) {
	var ruleName []byte

	if letter, err := p.parseLetter(); err != nil {
		return nil, err
	} else {
		ruleName = append(ruleName, letter)
	}

	for {
		if char, err := p.parseRuleChar(); err == nil {
			ruleName = append(ruleName, char)
		} else {
			break
		}
	}

	return ruleName, nil
}

func (p *SyntacticParser) parseRuleChar() (byte, error) {
	if letter, err := p.parseLetter(); err == nil {
		return letter, nil
	}

	if digit, err := p.parseDigit(); err == nil {
		return digit, nil
	}

	if hyphen, err := p.parseHyphen(); err == nil {
		return hyphen, nil
	}

	return byte(0), ErrUnexpectedChar
}

func (p *SyntacticParser) parseCharacter() (byte, error) {
	if letter, err := p.parseLetter(); err == nil {
		return letter, nil
	}

	if digit, err := p.parseDigit(); err == nil {
		return digit, nil
	}

	if symbol, err := p.parseSymbol(); err == nil {
		return symbol, nil
	}

	return byte(0), ErrUnexpectedChar
}

func (p *SyntacticParser) parseCharacterAndQuote() (byte, error) {
	if quote, err := p.parseQuote(); err == nil {
		return quote, err
	} else {
		return p.parseCharacter()
	}
}

func (p *SyntacticParser) parseCharacterAndDoubleQuote() (byte, error) {
	if quote, err := p.parseDoubleQuote(); err == nil {
		return quote, err
	} else {
		return p.parseCharacter()
	}
}

func (p *SyntacticParser) parseDefinitionSimbol() (*Token, error) {
	const name = "::="
	var token = Token{Name: []byte(name), Begin: p.pos, End: p.pos + 3}

	// Out of buffer check.
	if p.pos+len(name) >= len(p.buf) {
		return nil, io.EOF
	}

	// Is there expected characters.
	if string(p.buf[p.pos:p.pos+3]) != name {
		return nil, ErrUnexpectedChar
	} else {
		p.pos += 3
		return &token, nil
	}
}

func (p *SyntacticParser) parseDisjunction() (*Token, error) {
	if _, err := p.parseVerticalBar(); err != nil {
		return nil, err
	} else {
		return &Token{
			Name:  []byte{'|'},
			Begin: p.pos - 1,
			End:   p.pos,
		}, nil
	}
}

func (p *SyntacticParser) parseAtom() (Node, error) {
	var begin = p.pos

	// Parse terminal literal.
	if literal, err := p.parseLiteral(); err == nil {
		return &Terminal{Token{literal, begin, p.pos}}, nil
	}

	// Parse non-terminal.
	if nonTerminal, err := p.parseNonTerminal(); err == nil {
		return nonTerminal, nil
	} else {
		return nil, err
	}
}

func (p *SyntacticParser) parseLiteral() ([]byte, error) {
	if err := p.eof(); err != nil {
		return nil, err
	}

	var literal []byte

	switch p.buf[p.pos] {
	case '"': // Literals like "literal'sample".
		if _, err := p.parseDoubleQuote(); err != nil {
			return nil, err
		}

		for {
			if char, err := p.parseCharacterAndQuote(); err != nil {
				break
			} else {
				literal = append(literal, char)
			}
		}

		if _, err := p.parseDoubleQuote(); err != nil {
			return nil, NewDescError(err, p.pos, "terminal")
		}
	case '\'': // Literals like 'literal"sample'.
		if _, err := p.parseQuote(); err != nil {
			return nil, err
		}

		for {
			if char, err := p.parseCharacterAndDoubleQuote(); err != nil {
				break
			} else {
				literal = append(literal, char)
			}
		}

		if _, err := p.parseQuote(); err != nil {
			return nil, NewDescError(err, p.pos, "terminal")
		}
	default:
		return nil, NewDescError(ErrUnexpectedChar, p.pos, "terminal")
	}

	return literal, nil
}

func (p *SyntacticParser) parseNonTerminal() (Node, error) {
	var err error
	var token Token
	var begin = p.pos

	if _, err := p.parseLAngle(); err != nil {
		return nil, NewDescError(err, begin, "non-terminal")
	}

	if token.Name, err = p.parseRuleName(); err != nil {
		return nil, NewDescError(err, begin, "non-terminal")
	}

	if _, err := p.parseRAngle(); err != nil {
		return nil, NewDescError(err, begin, "non-terminal")
	}

	token.Begin = begin
	token.End = p.pos
	return &NonTerminal{token}, nil
}

func (p *SyntacticParser) parseLineEnd() error {
	if err := p.parseOptWhitespace(); err != nil {
		return err
	}

	if _, err := p.parseEOL(); err != nil {
		return err
	}

	var offset = p.pos
	if err := p.parseLineEnd(); err != nil {
		p.pos = offset
	}

	return nil
}

func (p *SyntacticParser) parseOptWhitespace() error {
	for p.pos < len(p.buf) {
		if p.buf[p.pos] == ' ' {
			p.pos++
		} else {
			break
		}
	}
	return nil
}

func (p *SyntacticParser) parseEOL() (byte, error) {
	return p.parseChar('\n')
}

func (p *SyntacticParser) parseLAngle() (byte, error) {
	return p.parseChar('<')
}

func (p *SyntacticParser) parseRAngle() (byte, error) {
	return p.parseChar('>')
}

func (p *SyntacticParser) parseHyphen() (byte, error) {
	return p.parseChar('-')
}

func (p *SyntacticParser) parseQuote() (byte, error) {
	return p.parseChar('\'')
}

func (p *SyntacticParser) parseDoubleQuote() (byte, error) {
	return p.parseChar('"')
}

func (p *SyntacticParser) parseVerticalBar() (byte, error) {
	return p.parseChar('|')
}

func (p *SyntacticParser) parseLetter() (byte, error) {
	if err := p.eof(); err != nil {
		return byte(0), err
	}

	var char = p.buf[p.pos]

	if (char >= 0x41 && char <= 0x5a) || (char >= 0x61 && char <= 0x7a) {
		p.pos++
		return char, nil
	} else {
		return byte(0), ErrUnexpectedChar
	}
}

func (p *SyntacticParser) parseDigit() (byte, error) {
	if err := p.eof(); err != nil {
		return byte(0), err
	}

	if char := p.buf[p.pos]; char >= 0x30 && char <= 0x39 {
		p.pos++
		return char, nil
	} else {
		return byte(0), ErrUnexpectedChar
	}
}

func (p *SyntacticParser) parseSymbol() (byte, error) {
	if err := p.eof(); err != nil {
		return byte(0), err
	}

	var char = p.buf[p.pos]
	var symbols = []byte{
		'|', ' ', '!', '#', '$', '%', '&', '(', ')', '*', '+', ',', '-', '.',
		'/', ':', ';', '>', '=', '<', '?', '@', '[', '\\', ']', '^', '_', '`',
		'{', '}', '~',
	}

	for _, symbol := range symbols {
		if symbol == char {
			p.pos++
			return char, nil
		}
	}

	return byte(0), ErrUnexpectedChar
}

func (p *SyntacticParser) parseChar(char byte) (byte, error) {
	if err := p.eof(); err != nil {
		return byte(0), err
	} else if p.buf[p.pos] != char {
		return byte(0), ErrUnexpectedChar
	} else {
		p.pos++
		return p.buf[p.pos-1], nil
	}
}
