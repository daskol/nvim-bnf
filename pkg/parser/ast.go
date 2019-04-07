package parser

import "strconv"

// VisitorFunc is a callback type for graph traversing. Its argument is the
// current node of traversing.
type VisitorFunc func(Node) error

// Node is a node of binary tree. In each node of a tree there is a token.
type Node interface {
	// Get the left child of a node.
	Left() Node
	// Get the right child of a node.
	Right() Node
}

// Token represents terminal, non-terminal, or any other lexemes. It is a base
// struct which is embedded below.
type Token struct {
	Name []byte
	// Begin encodes position where token begins. The possition is relative to
	// begin position of parent token.
	Begin int
	// End encodes position where token ends. The position is relateive as well
	// as in case of begin.
	End int
}

// Left does not return any node by default.
func (t *Token) Left() Node {
	return nil
}

// Right does not return any node by default.
func (t *Token) Right() Node {
	return nil
}

// String returns textual representation of token for debugging or something.
func (t *Token) String() string {
	return t.stringFromPositionAndName("Token")
}

func (t *Token) stringFromPositionAndName(token string) string {
	var name = string(t.Name)
	var pos = "begin=" + strconv.Itoa(t.Begin) + "; end=" + strconv.Itoa(t.End)
	return "<" + token + " name=" + name + "; " + pos + ">"
}

func (t *Token) stringFromPosition(token string) string {
	var pos = "begin=" + strconv.Itoa(t.Begin) + "; end=" + strconv.Itoa(t.End)
	return "<" + token + " " + pos + ">"
}

// Comment is a lexeme that encode comment in a source code.
type Comment struct {
	Token
}

func (c *Comment) Left() Node {
	return nil
}

func (c *Comment) Right() Node {
	return nil
}

// String returns textual representation of comment which could be usefull for
// debugging purposes.
func (c *Comment) String() string {
	return c.stringFromPosition("Comment")
}

// NonTerminal is a non-terminal symbols of input source.
type NonTerminal struct {
	Token
}

func (t *NonTerminal) String() string {
	return t.stringFromPositionAndName("NonTerminal")
}

// Terminal represents terminal symbols as nodes in an abstract syntax tree.
type Terminal struct {
	Token
}

func (t *Terminal) String() string {
	return t.stringFromPositionAndName("Terminal")
}

// Statement represents a BNF statement which could be empty (blank line) or
// not. In any case its right child points to comment. However, the left child
// is either nil or assignment expression.
type Statement struct {
	Rule    *AssignmentExpression
	Comment *Comment
}

func (s *Statement) Left() Node {
	return s.Rule
}

func (s *Statement) Right() Node {
	return s.Comment
}

func (s *Statement) String() string {
	return "<Statement>"
}

// Expression is a base type for any binary expression like disjunction of two
// compund rules or definition of compund rule itself.
type Expression struct {
	Token
	LeftChild  Node
	RightChild Node
}

func (e *Expression) Left() Node {
	return e.LeftChild
}

func (e *Expression) Right() Node {
	return e.RightChild
}

func (e *Expression) String() string {
	return "<Expression>"
}

// AlternativeExpression is a union of two or more right-hand sides of a
// production rule for the same left-hand side.
//
// The left child of the expression is one of CompoundExpression, NonTerminal,
// or Terminal. The right child could be one of AlternativeExpression,
// CompoundExpression, , NonTerminal, or Terminal.
type AlternativeExpression struct {
	Expression
}

func (e *AlternativeExpression) String() string {
	return e.stringFromPositionAndName("AlternativeExpression")
}

// AssignmentExpression assigns productions to a non-terminal symbol which is
// on the left from the assignment operator.
//
// The only possibility for type of the left child is a non-terminal symbol. On
// the right-hand side there is either AlternativeExpression,
// CompoundExpression, NonTerminal, or Terminal.
type AssignmentExpression struct {
	Expression
}

func (e *AssignmentExpression) String() string {
	return e.stringFromPositionAndName("AssignmentExpression")
}

// CompoundExpression combines two or more both terminals or non-terminals
// which represent right-hand side of a production rule. CompundExpression is
// designed in a way as it is in LISP-like languages. Since the expression is a
// list of Terminal or NonTerminal lexemmes, we could transform the list into
// binary tree which leaves are Terminals an NonTerminals.
//
// The left child of the compund expression could be either Terminal or
// NonTerminal. The right child is one of CompoundExpression, Terminal, or
// NonTerminal.
type CompoundExpression struct {
	Expression
}

func (e *CompoundExpression) String() string {
	return e.stringFromPosition("CompoundExpression")
}
