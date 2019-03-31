package parser

import (
	"bytes"
	"testing"
)

func TestSyntacticParser(t *testing.T) {
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
	})

	t.Run("BNF", func(t *testing.T) {
		var content = readBNFFile(t, "bnf.bnf")
		var parser = NewSyntacticParser(bytes.NewBuffer(content))
		var ast, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := ast.NoRules(); length != 18 {
			t.Errorf("too a few production rules: %d", length)
		}
	})
}
