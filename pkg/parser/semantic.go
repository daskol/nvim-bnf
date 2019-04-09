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

func (p *SemanticParser) Parse() (*AST, error) {
	if bytes, err := ioutil.ReadAll(p.Reader); err != nil {
		return nil, err
	} else {
		p.buf = bytes
		p.pos = 0
	}

	if rules, err := p.parseSyntax(); err != nil {
		return nil, &Error{err, p.pos + 1}
	} else {
		return &AST{rules: rules, semantic: true}, nil
	}
}

func (p *SemanticParser) eof() error {
	if p.pos == len(p.buf) {
		return io.EOF
	} else {
		return nil
	}
}

func (p *SemanticParser) parseSyntax() ([]*Statement, error) {
	var result []*Statement
	var stmt, err = p.parseRule()

	// Try to parse the first statement if there is any.
	switch {
	case err == io.EOF && stmt == nil:
		return nil, nil
	case err == io.EOF && stmt != nil:
		result = append(result, stmt)
		return result, nil
	case err != nil:
		return nil, err
	}

	// Recursively parse the rest.
	if stmts, err := p.parseSyntax(); err != nil {
		result = append(result, stmt)
	} else {
		result = append(result, stmt)
		result = append(result, stmts...)
	}

	return result, nil
}

func (p *SemanticParser) parseRule() (*Statement, error) {
	var err error
	var token *Token
	var expr = new(AssignmentExpression)
	var stmt = Statement{Rule: expr}

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if expr.LeftChild, err = p.parseNonTerminal(); err != nil {
		return nil, err
	}

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if token, err = p.parseDefinitionSimbol(); err != nil {
		return nil, err
	} else {
		expr.Token = *token
	}

	if err = p.parseOptWhitespace(); err != nil {
		return nil, err
	}

	if expr.RightChild, err = p.parseExpression(); err != nil {
		return nil, err
	}

	if err = p.parseLineEnd(); err == io.EOF {
		return &stmt, nil
	} else if err != nil {
		return nil, err
	}

	return &stmt, nil
}

func (p *SemanticParser) parseExpression() (Node, error) {
	var err error
	var offset int
	var root = new(AlternativeExpression)
	var token *Token

	// Parse single term list at first and back up position.
	if root.LeftChild, err = p.parseList(); err != nil {
		return nil, err
	} else {
		offset = p.pos
	}

	// Now try to parse multiple production rules.
	if err := p.parseOptWhitespace(); err != nil {
		p.pos = offset
		return root.LeftChild, nil
	}

	if token, err = p.parseDisjunction(); err != nil {
		p.pos = offset
		return root.LeftChild, nil
	} else {
		root.Token = *token
	}

	if err := p.parseOptWhitespace(); err != nil {
		p.pos = offset
		return root.LeftChild, nil
	}

	if root.RightChild, err = p.parseExpression(); err != nil {
		p.pos = offset
		return root.LeftChild, nil
	}

	return root, nil
}

func (p *SemanticParser) parseList() (Node, error) {
	var err error
	var offset = p.pos
	var root = new(CompoundExpression)
	var last = root
	var node Node

	// Use CompoundExpression to create the first element of lexemme list.
	if root.LeftChild, err = p.parseAtom(); err != nil {
		return nil, err
	}

	// Append CompoundExpression on each iteration.
	for {
		offset = p.pos

		if err := p.parseOptWhitespace(); err != nil {
			break
		}

		if node, err = p.parseAtom(); err != nil {
			break
		}

		var expr = &CompoundExpression{Expression{
			Token:      Token{Begin: offset, End: p.pos},
			LeftChild:  node,
			RightChild: nil,
		}}

		last.RightChild = expr
		last = expr
	}

	p.pos = offset

	// If there is only one lexeme then replace CompoundExpression with
	// Terminal or NonTerminal node.
	if last == root {
		return root.LeftChild, nil
	}

	// Otherwise, replace the last CompoundExpression with either Terminal or
	// NonTerminal node.
	var curr *CompoundExpression = root
	for curr.RightChild != last {
		curr = curr.RightChild.(*CompoundExpression)
	}

	// Uplift the last child to the previous one CompoundExpression.
	curr.RightChild = last.LeftChild
	return root, nil
}
