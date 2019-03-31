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
		var bnf, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := len(bnf.Rules); length != 7 {
			t.Errorf("too a few production rules: %d", length)
		}
	})

	t.Run("BNF", func(t *testing.T) {
		var content = readBNFFile(t, "bnf.bnf")
		var parser = NewSemanticParser(bytes.NewBuffer(content))
		var bnf, err = parser.Parse()

		if err != nil {
			t.Fatalf("failed to parse grammar: %s", err)
		}

		if length := len(bnf.Rules); length != 18 {
			t.Errorf("too a few production rules: %d", length)
		}
	})
}
