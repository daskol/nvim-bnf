// Tool nvim-bnf implement NeoVim MsgPack RPC and provides semantic syntax
// hightlighter for BNF.
package main

import (
	"errors"
	"log"
	"runtime/debug"

	"github.com/daskol/nvim-bnf/bnf"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
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
		var begin, end int

		switch node := node.(type) {
		case *bnf.ProductionRule:
			grp = "Operator"
			begin = node.Begin
			end = node.End
			batch.AddBufferHighlight(buf, 0, grp, row, begin, end, nil)
		case *bnf.Token:
			grp = "Identifier"
			begin = node.Begin
			end = node.End
			batch.AddBufferHighlight(buf, 0, grp, row, begin, end, nil)
		case *bnf.Stmt:
			grp = "Operator"
			begin = node.Begin
			end = node.End
			batch.AddBufferHighlight(buf, 0, grp, row, begin, end, nil)
		case bnf.List:
			for _, token := range node {
				if token.Terminal {
					grp = "String"
				} else {
					grp = "Identifier"
				}
				begin = token.Begin
				end = token.End
				batch.AddBufferHighlight(buf, 0, grp, row, begin, end, nil)
			}
		default:
			logger.Warnf("visiting unexpected token: %T", node)
		}

		return nil
	})
}

func HightlightDoc(v *nvim.Nvim, buf nvim.Buffer) {
	if doc, ok := DocIndex[buf]; ok {
		doc.Hightlight(v, buf)
	}
}

func HightlightHunk(v *nvim.Nvim, buf nvim.Buffer, from, to int) {
	if doc, ok := DocIndex[buf]; ok {
		doc.HightlightHunk(v, buf, from, to)
	}
}

func HandleBufReadEvent(v *nvim.Nvim, filename *string) {
	logger.Debugf("HandleBufReadEvent(%s)", *filename)
	var buf, err = v.CurrentBuffer()

	if err != nil {
		logger.Errorf("failed to get current buffer: %s", err)
		return
	}

	if err = AttachToBuffer(v, &buf); err != nil {
		logger.Errorf("failed to attach to buffer: %s", err)
	}

	logger.Infof("buffer %d was attached to plugin", buf)
}

func HandleBufLinesEvent(
	v *nvim.Nvim, buf *nvim.Buffer, changedTick int, firstLine, lastLine int,
	data [][]byte, more bool,
) {
	logger.Debugf(
		"HandleBufLinesEvent(%s, %d, %d, %d, [[...]], %t)",
		buf, changedTick, firstLine, lastLine, more,
	)

	if lastLine == -1 {
		DocIndex[*buf] = &Document{Lines: data}
		HightlightDoc(v, *buf)
	} else {
		var doc, ok = DocIndex[*buf]

		if !ok {
			logger.Warnf("unknown buffer: %s", buf)
			return
		}

		var from, to = doc.Update(data, firstLine, lastLine)
		doc.HightlightHunk(v, *buf, from, to)
	}
}

func HandleBufDetachEvent(v *nvim.Nvim, buf *nvim.Buffer) {
	logger.Debugf("HandleBufDetachEvent(%s)", buf)

	if err := DetachFromBuffer(v, buf); err != nil {
		logger.Errorf("failed to detac buffer: %s", err)
		return
	}

	logger.Infof("buffer %d was detached from plugin", buf)
}

func HandleBufChangedTickEvent(
	v *nvim.Nvim, buf *nvim.Buffer, changedTick int,
) {
	logger.Debugf("HandleBufChangedTickEvent(%s, %d)", buf, changedTick)
}

func main() {
	if ptr, err := NewLogger(); err != nil {
		log.Fatalf("failed to instantiate logger: %s", err)
		return
	} else {
		logger = ptr
	}

	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("error occured during logger closing: %s", err)
		}
	}()

	logger.Infof("register plugin handlers")
	plugin.Main(func(p *plugin.Plugin) error {
		// Register event handlers during actual loading.
		if p.Nvim != nil {
			p.Nvim.RegisterHandler("nvim_buf_changedtick_event", HandleBufChangedTickEvent)
			p.Nvim.RegisterHandler("nvim_buf_detach_event", HandleBufDetachEvent)
			p.Nvim.RegisterHandler("nvim_buf_lines_event", HandleBufLinesEvent)
		}

		// Register autocommands.
		for _, event := range []string{"BufRead", "BufNewFile"} {
			var opts = &plugin.AutocmdOptions{
				Event:   event,
				Group:   "nvim-bnf",
				Pattern: "*.bnf",
				Eval:    `expand("<afile>")`,
			}
			p.HandleAutocmd(opts, HandleBufReadEvent)
		}

		return nil
	})
}
