package bnf

import (
	"io"
	"io/ioutil"
)

type BNFParser struct {
	Reader io.Reader

	buf []byte
	pos int
}

func (p *BNFParser) Parse() (*BNF, error) {
	if bytes, err := ioutil.ReadAll(p.Reader); err != nil {
		return nil, err
	} else {
		p.buf = bytes
		p.pos = 0
	}

	if rules, err := p.parseSyntax(); err != nil {
		return nil, &Error{err, p.pos + 1}
	} else {
		return &BNF{rules}, nil
	}
}

func (p *BNFParser) eof() error {
	if p.pos == len(p.buf) {
		return io.EOF
	} else {
		return nil
	}
}

func (p *BNFParser) parseSyntax() ([]*ProductionRule, error) {
	var err error
	var ret []*ProductionRule
	var rule *ProductionRule

	if rule, err = p.parseRule(); err != nil {
		return nil, err
	}

	if rules, err := p.parseSyntax(); err != nil {
		ret = append(ret, rule)
	} else {
		ret = append(ret, rule)
		ret = append(ret, rules...)
	}

	return ret, nil
}

func (p *BNFParser) parseRule() (*ProductionRule, error) {
	var err error
	var rule ProductionRule

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if _, err = p.parseLAngle(); err != nil {
		return nil, err
	}

	if rule.Name, err = p.parseRuleName(); err != nil {
		return nil, err
	}

	if _, err = p.parseRAngle(); err != nil {
		return nil, err
	}

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if err = p.parseDefinitionSimbol(); err != nil {
		return nil, err
	}

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if rule.Expressions, err = p.parseExpression(); err != nil {
		return nil, err
	}

	if err = p.parseLineEnd(); err != nil {
		return nil, err
	}

	return &rule, nil
}

func (p *BNFParser) parseRuleName() ([]byte, error) {
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

func (p *BNFParser) parseRuleChar() (byte, error) {
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

func (p *BNFParser) parseCharacter() (byte, error) {
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

func (p *BNFParser) parseCharacterAndQuote() (byte, error) {
	if quote, err := p.parseQuote(); err == nil {
		return quote, err
	} else {
		return p.parseCharacter()
	}
}

func (p *BNFParser) parseCharacterAndDoubleQuote() (byte, error) {
	if quote, err := p.parseDoubleQuote(); err == nil {
		return quote, err
	} else {
		return p.parseCharacter()
	}
}

func (p *BNFParser) parseDefinitionSimbol() error {
	if string(p.buf[p.pos:p.pos+3]) != "::=" {
		return ErrUnexpectedChar
	} else {
		p.pos += 3
		return nil
	}
}

func (p *BNFParser) parseExpression() ([]Expression, error) {
	var ret []Expression
	var offset int

	// Parse single term list at first and back up position.
	if terms, err := p.parseList(); err != nil {
		return nil, err
	} else {
		offset = p.pos
		ret = append(ret, terms)
	}

	// Now try to parse multiple production rules.
	if err := p.parseOptWhitespace(); err != nil {
		p.pos = offset
		return ret, nil
	}

	if _, err := p.parseVerticalBar(); err != nil {
		p.pos = offset
		return ret, nil
	}

	if err := p.parseOptWhitespace(); err != nil {
		p.pos = offset
		return ret, nil
	}

	if exprs, err := p.parseExpression(); err != nil {
		p.pos = offset
		return ret, nil
	} else {
		ret = append(ret, exprs...)
	}

	return ret, nil
}

func (p *BNFParser) parseList() (Expression, error) {
	var offset = p.pos
	var terms []*Term

	if term, err := p.parseTerm(); err != nil {
		return nil, err
	} else {
		terms = append(terms, term)
	}

	for {
		offset = p.pos

		if err := p.parseOptWhitespace(); err != nil {
			break
		}

		if term, err := p.parseTerm(); err != nil {
			break
		} else {
			terms = append(terms, term)
		}
	}

	p.pos = offset
	return terms, nil
}

func (p *BNFParser) parseTerm() (*Term, error) {
	// Parse terminal literal.
	if literal, err := p.parseLiteral(); err == nil {
		return &Term{literal, true}, nil
	}

	// Parse non-terminal.
	if _, err := p.parseLAngle(); err != nil {
		return nil, err
	}

	var term Term

	if ruleName, err := p.parseRuleName(); err != nil {
		return nil, err
	} else {
		term.Name = ruleName
	}

	if _, err := p.parseRAngle(); err != nil {
		return nil, err
	}

	return &term, nil
}

func (p *BNFParser) parseLiteral() ([]byte, error) {
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
			return nil, err
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
			return nil, err
		}
	default:
		return nil, ErrUnexpectedChar
	}

	return literal, nil
}

func (p *BNFParser) parseLineEnd() error {
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

func (p *BNFParser) parseOptWhitespace() error {
	for p.pos < len(p.buf) {
		if p.buf[p.pos] == ' ' {
			p.pos++
		} else {
			break
		}
	}
	return nil
}

func (p *BNFParser) parseEOL() (byte, error) {
	return p.parseChar('\n')
}

func (p *BNFParser) parseLAngle() (byte, error) {
	return p.parseChar('<')
}

func (p *BNFParser) parseRAngle() (byte, error) {
	return p.parseChar('>')
}

func (p *BNFParser) parseHyphen() (byte, error) {
	return p.parseChar('-')
}

func (p *BNFParser) parseQuote() (byte, error) {
	return p.parseChar('\'')
}

func (p *BNFParser) parseDoubleQuote() (byte, error) {
	return p.parseChar('"')
}

func (p *BNFParser) parseVerticalBar() (byte, error) {
	return p.parseChar('|')
}

func (p *BNFParser) parseLetter() (byte, error) {
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

func (p *BNFParser) parseDigit() (byte, error) {
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

func (p *BNFParser) parseSymbol() (byte, error) {
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

func (p *BNFParser) parseChar(char byte) (byte, error) {
	if err := p.eof(); err != nil {
		return byte(0), err
	} else if p.buf[p.pos] != char {
		return byte(0), ErrUnexpectedChar
	} else {
		p.pos++
		return p.buf[p.pos-1], nil
	}
}
