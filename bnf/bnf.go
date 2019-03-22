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

type Token = Term

// Term represents terminal or non-terminal.
type Term struct {
	Name     []byte
	Terminal bool
	Begin    int
	End      int
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

// Expression is a list of terminals and non-terminals.
type Expression []*Term

func (e *Expression) String() string {
	var parts []string
	for _, term := range *e {
		parts = append(parts, term.String())
	}
	return strings.Join(parts, " ")
}

// ProductionRule is a production rule itself. Actually, it contains several
// rules for a non-terminal.
type ProductionRule struct {
	Token       // Points to lexeme that contains `::=`.
	Name        *Term
	Expressions []Expression
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
