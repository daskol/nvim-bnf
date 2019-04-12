package highlighting

import (
	"errors"
	"runtime/debug"

	"github.com/daskol/nvim-bnf/pkg/parser"
	"github.com/neovim/go-client/nvim"
)

var DocIndex = make(map[nvim.Buffer]*Document)
var NonTerminalIndex = make(map[string]uint)

// Document is a mirrored content of NeoVim buffer. This object provides
// human-readable interface for document management, hightlighting and
// versioning.
type Document struct {
	Lines [][]byte

	batch  *nvim.Batch
	buffer *nvim.Buffer
}

// Get returns line in document if it exists.
func (d *Document) Get(idx int) ([]byte, bool) {
	if idx < 0 || idx >= len(d.Lines) {
		return nil, false
	} else {
		return d.Lines[idx], true
	}
}

// NoLines returns number of lines in document.
func (d *Document) NoLines() int {
	return len(d.Lines)
}

// Update updates document with a hunk of lines.
func (d *Document) Update(lines [][]byte, from, to int) (int, int) {
	var nolines = len(lines)
	var firstLines = d.Lines[:from]
	var lastLines [][]byte

	if to < len(d.Lines) {
		lastLines = d.Lines[to:]
	}

	lines = append(firstLines, lines...)
	lines = append(lines, lastLines...)
	d.Lines = lines
	return from, from + nolines
}

// Hightlight adds hightlight to buffer for an entire document.
func (d *Document) Hightlight(v *nvim.Nvim, buf nvim.Buffer) {
	d.HightlightHunk(v, buf, 0, d.NoLines())
}

// HightlightHunk adds hightlight to a chunk of lines of a buffer.
func (d *Document) HightlightHunk(v *nvim.Nvim, buf nvim.Buffer, from, to int) {
	if from < 0 {
		from = 0
	}

	if to > d.NoLines() {
		to = d.NoLines()
	}

	logger.Debugf("hightlight hunk from %d to %d", from, to)
	var batch = v.NewBatch()

	for line := from; line != to; line++ {
		var ast, err = d.parse(d.Lines[line])

		// Skip the line if it causes parsing errors.
		if err != nil {
			continue
		}

		// Update completion index.
		if err = d.updateCompletionIndex(ast); err != nil {
			logger.Warnf("failed to update completion index: %s", err)
		}

		// Hightlight line.
		if err = d.hightlightLine(batch, buf, line, ast); err != nil {
			logger.Warnf(
				"failed to hightlight line %d of %s: %s",
				line, buf, err,
			)
		}
	}

	if err := batch.Execute(); err != nil {
		logger.Errorf("failed to execute batch RPC call: %s", err)
	}
}

func (d *Document) parse(line []byte) (*parser.AST, error) {
	var ast *parser.AST
	var err error
	defer func() {
		// TODO(@daskol): Test parser heavily!
		if ctx := recover(); ctx != nil {
			logger.Errorf("recovery: %s\n%s", ctx, debug.Stack())
			err = errors.New("recovery during parsing")
		}
	}()

	if ast, err = parser.Parse(line); err != nil {
		logger.Warnf("failed to parse: %s", err)
		return nil, err
	} else {
		return ast, nil
	}
}

func (d *Document) hightlightLine(
	batch *nvim.Batch,
	buf nvim.Buffer,
	row int,
	ast *parser.AST,
) error {
	batch.ClearBufferHighlight(buf, -1, row, row+1)

	// Traverse abstract tree and hightlight lexemes.
	var nonodes, err = ast.Traverse(func(node parser.Node) error {
		var grp string
		var begin, end, res int

		switch node := node.(type) {
		case *parser.AssignmentExpression:
			grp = "Operator"
			begin = node.Begin
			end = node.End
		case *parser.Terminal:
			grp = "String"
			begin = node.Begin
			end = node.End
		case *parser.NonTerminal:
			grp = "Identifier"
			begin = node.Begin
			end = node.End
		case *parser.AlternativeExpression:
			grp = "Operator"
			begin = node.Begin
			end = node.End
		case *parser.Comment:
			grp = "Comment"
		default:
			return nil
		}

		batch.AddBufferHighlight(buf, 0, grp, row, begin, end, &res)

		return nil
	})

	// If error was occured during traversing then exit.
	if err != nil {
		return err
	}

	// If there is no lexemes in tree then exit as well.
	if nonodes == 0 {
		return nil
	}

	// Update virtual text with error annotation.
	if err := ast.Error(); err != nil {
		var res = 0
		var text = "syn: " + err.Error()
		if err, ok := err.(*parser.DescError); ok {
			text = err.String()
		}
		var chunks = []Chunk{NewChunk(text, "Error")}
		SetVirtualText(batch, &buf, 0, row, chunks, NoOpts, &res)
	}

	return nil
}

func (d *Document) updateCompletionIndex(ast *parser.AST) error {
	var _, err = ast.Traverse(func(node parser.Node) error {
		if node, ok := node.(*parser.NonTerminal); ok {
			var counter = NonTerminalIndex[string(node.Name)]
			NonTerminalIndex[string(node.Name)] = counter + 1
		}

		return nil
	})
	return err
}
