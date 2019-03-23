// Package bnf contains parser for BNF metalanguage.
package bnf

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

var ErrNotImplemented = errors.New("bnf: not implemented")
var ErrUnexpectedChar = errors.New("bnf: unexpected character")

// Error represent parsing error and contains some context information.
type Error struct {
	err error
	pos int
}

func (e *Error) Error() string {
	return e.err.Error() + " at position " + strconv.Itoa(e.pos)
}

// Node is a node of binary tree. In each node of a tree there is a token.
type Node interface {
	// Get the left child of a node.
	Left() Node
	// Get the right child of a node.
	Right() Node
	// Get value which is assigned to a node.
	Value() Node
}

// VisitorFunc is a callback type for graph traversing. Its argument is the
// current node of traversing.
type VisitorFunc func(Node) error

// Visit implements in-order graph traversal procedure.
func Visit(root Node, f VisitorFunc) error {
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

type Token = Term

// Term represents terminal or non-terminal.
type Term struct {
	Name     []byte
	Terminal bool
	// Begin encodes position where token begins. The possition is relative to
	// begin position of parent token.
	Begin int
	// End encodes position where token ends. The position is relateive as well
	// as in case of begin.
	End int
}

func (t *Term) Left() Node {
	return nil
}

func (t *Term) Right() Node {
	return nil
}

func (t *Term) Value() Node {
	return t
}

func (t *Term) String() string {
	var name = string(t.Name)
	var pos = "begin=" + strconv.Itoa(t.Begin) + "; end=" + strconv.Itoa(t.End)
	var terminal = "false"
	if t.Terminal {
		terminal = "true"
	}
	return "<Term name=" + name + "; terminal=" + terminal + "; " + pos + ">"
}

type List = Expression

// Expression is a list of terminals and non-terminals.
type Expression []*Term

func (e Expression) Left() Node {
	return nil
}

func (e Expression) Right() Node {
	return nil
}

func (e Expression) Value() Node {
	return e
}

func (e *Expression) String() string {
	var parts []string
	for _, term := range *e {
		parts = append(parts, term.String())
	}
	return strings.Join(parts, " ")
}

type Stmt struct {
	Token // Points to lexeme that contains '|`.
	Head  List
	Tail  Node
}

func (s *Stmt) Left() Node {
	return s.Head
}

func (s *Stmt) Right() Node {
	return s.Tail
}

func (s *Stmt) Value() Node {
	return s
}

// ProductionRule is a production rule itself. Actually, it contains several
// rules for a non-terminal.
type ProductionRule struct {
	Token       // Points to lexeme that contains `::=`.
	Name        *Term
	Stmt        Node
	Expressions []Expression
}

func (r *ProductionRule) Left() Node {
	return r.Name
}

func (r *ProductionRule) Right() Node {
	return r.Stmt
}

func (r *ProductionRule) Value() Node {
	return r
}

func (r *ProductionRule) String() string {
	var parts []string
	for _, expr := range r.Expressions {
		parts = append(parts, expr.String())
	}
	return r.Name.String() + " -> " + strings.Join(parts, " | ")
}

// BNF types corresponds parsed BNF grammar.
type BNF struct {
	Rules []*ProductionRule
}

func (bnf *BNF) String() string {
	var rules []string
	for _, rule := range bnf.Rules {
		rules = append(rules, rule.String())
	}
	return strings.Join(rules, "\n")
}

// Parse parses BNF grammar.
func Parse(source []byte) (*BNF, error) {
	return (&BNFParser{
		Reader: bytes.NewBuffer(source),
	}).Parse()
}
