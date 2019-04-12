package parser

import (
	"bytes"
	"testing"
)

func TestSyntacticParser(t *testing.T) {
	t.Run("EmptySource", func(t *testing.T) {
		var content = []byte("")
		var parser = NewSyntacticParser(bytes.NewBuffer(content))
		var ast, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := ast.NoRules(); length != 0 {
			t.Errorf("too a few production rules: %d", length)
		}

		defer func() {
			if ctx := recover(); ctx != nil {
				t.Fatalf("failed to traverse syntax tree")
			}
		}()

		ast.traverseSyntacticTree(func(node Node) error {
			return nil
		})
	})

	t.Run("US Postal Address", func(t *testing.T) {
		var content = readBNFFile(t, "us-postal-address.bnf")
		var parser = NewSyntacticParser(bytes.NewBuffer(content))
		var ast, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := ast.NoRules(); length != 7 {
			t.Errorf("too a few production rules: %d", length)
		}

		// Let's check the third statement which is
		//   <personal-part> ::= <initial> "." | <first-name>
		var lemmes = ast.lemmes[2]

		if numb := len(lemmes); numb != 6 {
			t.Fatalf("too few number of lexemes in statement: %d", numb)
		}

		if nonterm, ok := lemmes[0].(*NonTerminal); !ok {
			t.Errorf("wrong type of rule name: %T", lemmes[0])
		} else if name := string(nonterm.Name); name != "personal-part" {
			t.Errorf("wrong rule name: %s", name)
		}

		if _, ok := lemmes[1].(*AssignmentExpression); !ok {
			t.Errorf("wrong type of assignement expression: %T", lemmes[1])
		}

		if term, ok := lemmes[3].(*Terminal); !ok {
			t.Errorf("wrong type of terminal: %T", lemmes[3])
		} else if name := string(term.Name); name != "." {
			t.Errorf("wrong name of terminal: %s", name)
		}

		if _, ok := lemmes[4].(*AlternativeExpression); !ok {
			t.Errorf("wrong type of alternative expression: %T", lemmes[4])
		}
	})

	t.Run("BNF", func(t *testing.T) {
		var content = readBNFFile(t, "bnf.bnf")
		var parser = NewSyntacticParser(bytes.NewBuffer(content))
		var ast, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := ast.NoRules(); length != 19 {
			t.Errorf("too a few production rules: %d", length)
		}
	})
}
