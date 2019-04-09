package parser

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func readBNFFile(t *testing.T, filename string) []byte {
	var bytes, err = ioutil.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	return bytes
}

func TestSemanticParser(t *testing.T) {
	t.Run("US Postal Address", func(t *testing.T) {
		var content = readBNFFile(t, "us-postal-address.bnf")
		var parser = NewSemanticParser(bytes.NewBuffer(content))
		var ast, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := ast.NoRules(); length != 7 {
			t.Errorf("too a few production rules: %d", length)
		}

		// Let's check the third statement which is
		//   <personal-part> ::= <initial> "." | <first-name>
		var stmt = ast.rules[2]
		var rule = stmt.Rule

		if stmt.Comment != nil {
			t.Errorf("statement does not contains comment")
		}

		if rule == nil {
			t.Fatalf("there is no parsed rule")
		}

		// Check the left-hand side of assignment expression.
		if node := rule.Left(); node == nil {
			t.Fatalf("lhs of assignment expression is nil")
		} else if lhs, ok := node.(*NonTerminal); !ok {
			t.Fatalf("wrong type of lhs of assignment expression: %T", node)
		} else if name := string(lhs.Name); name != "personal-part" {
			t.Fatalf("wrong rule name: %s", name)
		}

		// Check the right-hand side of assignment expression.
		var left, right Node

		if node := rule.Right(); node == nil {
			t.Fatalf("rhs of assignment expression is nil")
		} else if rhs, ok := node.(*AlternativeExpression); !ok {
			t.Fatalf("wrong type of rhs of assignment expression: %T", node)
		} else {
			left = rhs.Left()
			right = rhs.Right()
		}

		// Check the first alternative expression.
		if alt1, ok := left.(*CompoundExpression); !ok {
			t.Fatalf("wrong type of the first alternative: %T", left)
		} else if lex1 := alt1.Left(); lex1 == nil {
			t.Fatalf("element of compound expression is missing")
		} else if nonTerm, ok := lex1.(*NonTerminal); !ok {
			t.Fatalf("wrong type of the first lexeme: %T", lex1)
		} else if name := string(nonTerm.Name); name != "initial" {
			t.Fatalf("wrong name of the first lexeme: %s", name)
		} else if lex2 := alt1.Right(); lex2 == nil {
			t.Fatalf("element of compound expression is missing")
		} else if term, ok := lex2.(*Terminal); !ok {
			t.Fatalf("wrong type of the second lexeme: %T", lex2)
		} else if name := string(term.Name); name != "." {
			t.Fatalf("wrong terminal: %s", name)
		}

		// Check the second alternative expression.
		if alt2, ok := right.(*NonTerminal); !ok {
			t.Fatalf("wrong type of the second alternative: %T", right)
		} else if name := string(alt2.Name); name != "first-name" {
			t.Errorf("wrong token name: %s", name)
		}
	})

	t.Run("BNF", func(t *testing.T) {
		var content = readBNFFile(t, "bnf.bnf")
		var parser = NewSemanticParser(bytes.NewBuffer(content))
		var ast, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := ast.NoRules(); length != 19 {
			t.Errorf("too a few production rules: %d", length)
		}
	})
}
