package parser

import (
	"io"
	"io/ioutil"
)

// SemanticParser performs semantical parsing of the input according to grammar
// of BNF.
type SemanticParser struct {
	SyntacticParser
}

func NewSemanticParser(reader io.Reader) *SemanticParser {
	return &SemanticParser{SyntacticParser: *NewSyntacticParser(reader)}
}

func (p *SemanticParser) Parse() (*BNF, error) {
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

func (p *SemanticParser) eof() error {
	if p.pos == len(p.buf) {
		return io.EOF
	} else {
		return nil
	}
}

func (p *SemanticParser) parseSyntax() ([]*ProductionRule, error) {
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

func (p *SemanticParser) parseRule() (*ProductionRule, error) {
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
