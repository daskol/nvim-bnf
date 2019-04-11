// Package bnf contains parser for BNF metalanguage.
package parser

import (
	"bytes"
	"io"
	"strconv"
)

// AST type corresponds parsed BNF grammar. We use the same AST type for both
// semantic parse tree and syntactic parse tree (which is actually a list of
// lists).
type AST struct {
	// Save the parsing error.
	err error
	// List of lists of terms. Each list corresponds to each line of the
	// source.
	lemmes [][]Node
	// List of production rules. Each production rule is a line in the source.
	rules []*Statement
	// True if the AST was produced be semantic parser.
	semantic bool
}

// Error provides access to saved semantic parsing errors.
func (ast *AST) Error() error {
	return ast.err
}

// NoRules gets the number of parsed rules.
func (ast *AST) NoRules() int {
	if ast.semantic {
		return len(ast.rules)
	} else {
		return len(ast.lemmes)
	}
}

// String returns textua representation of an object.
func (ast *AST) String() string {
	var norules = ast.NoRules()
	return "<AST norules=" + strconv.Itoa(norules) + ";>"
}

// Traverse implements in-order graph traversal procedure.
func (ast *AST) Traverse(visitor VisitorFunc) error {
	if ast.semantic {
		return ast.traverseSemanticTree(visitor)
	} else {
		return ast.traverseSyntacticTree(visitor)
	}
}

func (ast *AST) traverseSemanticTree(visitor VisitorFunc) error {
	// TODO(@daskol): Remove this tests in the future!
	if len(ast.rules) == 0 {
		return ErrNoStatements
	} else if ast.rules[0] == nil {
		return ErrEmptyRule
	} else {
		return ast.visit(ast.rules[0], visitor)
	}
}

func (ast *AST) traverseSyntacticTree(visitor VisitorFunc) error {
	for _, node := range ast.lemmes[0] {
		if err := visitor(node); err != nil {
			return err
		}
	}
	return nil
}

func (ast *AST) visit(root Node, f VisitorFunc) error {
	var stack = []Node{root}

	for len(stack) != 0 {
		var last = len(stack) - 1
		var node = stack[last]

		// If the last node in the stack is nil it means that we should put
		// right node to stack. Otherwise, put left node and continue
		// iterating.
		if node != nil {
			stack = append(stack, node.Left())
			continue
		}

		// Pop last (nil) node.
		stack = stack[:last]
		last = len(stack) - 1

		if last+1 == 0 {
			break
		}

		// Pop parent node (of nil node) and push right node.
		node = stack[last]
		stack[last] = node.Right()

		// Call visitor callback on the parent node.
		if err := f(node); err != nil {
			return err
		}
	}

	return nil
}

// Parse parses BNF grammar.
func Parse(source []byte) (*AST, error) {
	var origin bytes.Buffer
	var replica = io.TeeReader(bytes.NewBuffer(source), &origin)
	var astSem, errSem = NewSemanticParser(replica).Parse()

	if errSem == nil {
		return astSem, nil
	}

	// Fallback to syntactic parser on error.
	var astSyn, errSyn = NewSyntacticParser(&origin).Parse()

	if errSyn != nil {
		return nil, errSyn
	} else {
		astSyn.err = errSem
	}

	return astSyn, nil
}
