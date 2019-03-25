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

	if rule, err = p.parseRule(); err == io.EOF {
		ret = append(ret, rule)
		return ret, nil
	} else if err != nil {
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
	var token *Token
	var rule = new(ProductionRule)

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if rule.Name, err = p.parseNonTerminal(); err != nil {
		return nil, err
	}

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if token, err = p.parseDefinitionSimbol(); err != nil {
		return nil, err
	} else {
		rule.Token = *token
	}

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if rule.Stmt, err = p.parseExpression(); err != nil {
		return nil, err
	}

	if err = p.parseLineEnd(); err == io.EOF {
		return rule, nil
	} else if err != nil {
		return nil, err
	}

	return rule, nil
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

func (p *BNFParser) parseDefinitionSimbol() (*Token, error) {
	const name = "::="
	var token = Token{Name: []byte(name), Begin: p.pos, End: p.pos + 3}

	if string(p.buf[p.pos:p.pos+3]) != name {
		return nil, ErrUnexpectedChar
	} else {
		p.pos += 3
		return &token, nil
	}
}

func (p *BNFParser) parseDisjunction() (*Token, error) {
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

func (p *BNFParser) parseExpression() (Node, error) {
	var err error
	var offset int
	var ret []Expression
	var stmt Stmt
	var head List
	var tail Node
	var token *Token

	// Parse single term list at first and back up position.
	if head, err = p.parseList(); err != nil {
		return nil, err
	} else {
		offset = p.pos
		ret = append(ret, head)
	}

	// Now try to parse multiple production rules.
	if err := p.parseOptWhitespace(); err != nil {
		p.pos = offset
		return head, nil
	}

	if token, err = p.parseDisjunction(); err != nil {
		p.pos = offset
		return head, nil
	} else {
		stmt.Token = *token
	}

	if err := p.parseOptWhitespace(); err != nil {
		p.pos = offset
		return head, nil
	}

	if tail, err = p.parseExpression(); err != nil {
		p.pos = offset
		return head, nil
	}

	stmt.Head = head
	stmt.Tail = tail
	return &stmt, nil
}

func (p *BNFParser) parseList() (List, error) {
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
	var begin = p.pos

	// Parse terminal literal.
	if literal, err := p.parseLiteral(); err == nil {
		return &Term{literal, true, begin, p.pos}, nil
	}

	// Parse non-terminal.
	if nonTerminal, err := p.parseNonTerminal(); err == nil {
		return nonTerminal, nil
	} else {
		return nil, err
	}
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

func (p *BNFParser) parseNonTerminal() (*Term, error) {
	var err error
	var nonTerminal = new(Term)
	var begin = p.pos

	if _, err := p.parseLAngle(); err != nil {
		return nil, err
	}

	if nonTerminal.Name, err = p.parseRuleName(); err != nil {
		return nil, err
	}

	if _, err := p.parseRAngle(); err != nil {
		return nil, err
	}

	nonTerminal.Begin = begin
	nonTerminal.End = p.pos
	return nonTerminal, nil
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