package main

import (
	"errors"
	"runtime/debug"

	"github.com/daskol/nvim-bnf/bnf"
	"github.com/neovim/go-client/nvim"
)

var DocIndex = make(map[nvim.Buffer]*Document)

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
	var lastLines = d.Lines[to:]

	lines = append(firstLines, lines...)
	lines = append(lines, lastLines...)
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

		switch {
		case err != nil:
			var res = 0
			var text = "parsing error: " + err.Error()
			var chunks = []Chunk{NewChunk(text, "Error")}
			SetVirtualText(batch, &buf, 0, line, chunks, NoOpts, &res)
		case err == nil:
			if err = d.hightlightLine(batch, buf, line, ast); err != nil {
				logger.Warnf(
					"failed to hightlight line %d of %s: %s",
					line, buf, err,
				)
			}
		}
	}

	if err := batch.Execute(); err != nil {
		logger.Errorf("failed to execute batch RPC call: %s", err)
	}
}

func (d *Document) parse(line []byte) (*bnf.BNF, error) {
	var ast *bnf.BNF
	var err error
	defer func() {
		// TODO(@daskol): Test parser heavily!
		if ctx := recover(); ctx != nil {
			logger.Errorf("recovery: %s\n%s", ctx, debug.Stack)
			err = errors.New("recovery during parsing")
		}
	}()

	// TODO(@daskol): Make more extensive parser tests!
	if ast, err = bnf.Parse(line); err != nil {
		logger.Warnf("failed to parse: %s", err)
		return nil, err
	} else if len(ast.Rules) == 0 {
		return nil, errors.New("nvim-bnf: there is no productions")
	} else if ast.Rules[0] == nil {
		return nil, errors.New("nvim-bnf: rule is empty")
	} else {
		return ast, nil
	}
}

func (d *Document) hightlightLine(
	batch *nvim.Batch,
	buf nvim.Buffer,
	row int,
	ast *bnf.BNF,
) error {
	batch.ClearBufferHighlight(buf, -1, row, row+1)
	return bnf.Visit(ast.Rules[0], func(node bnf.Node) error {
		var grp string
		var begin, end, res int

		switch node := node.(type) {
		case *bnf.ProductionRule:
			grp = "Operator"
			begin = node.Begin
			end = node.End
			batch.AddBufferHighlight(buf, 0, grp, row, begin, end, &res)
		case *bnf.Token:
			grp = "Identifier"
			begin = node.Begin
			end = node.End
			batch.AddBufferHighlight(buf, 0, grp, row, begin, end, &res)
		case *bnf.Stmt:
			grp = "Operator"
			begin = node.Begin
			end = node.End
			batch.AddBufferHighlight(buf, 0, grp, row, begin, end, &res)
		case bnf.List:
			for _, token := range node {
				if token.Terminal {
					grp = "String"
				} else {
					grp = "Identifier"
				}
				begin = token.Begin
				end = token.End
				batch.AddBufferHighlight(buf, 0, grp, row, begin, end, &res)
			}
		default:
			logger.Warnf("visiting unexpected token: %T", node)
		}

		return nil
	})
}
