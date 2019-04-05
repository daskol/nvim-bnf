package parser

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

// AlternativeExpression is a union of two or more right-hand sides of a
// production rule for the same left-hand side.
//
// The left child of the expression is one of CompoundExpression, NonTerminal,
// or Terminal. The right child could be one of AlternativeExpression,
// CompoundExpression, , NonTerminal, or Terminal.
type AlternativeExpression struct {
	Expression
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
