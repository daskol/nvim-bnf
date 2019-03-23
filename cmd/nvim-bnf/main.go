// Tool nvim-bnf implement NeoVim MsgPack RPC and provides semantic syntax
// hightlighter for BNF.
package main

import (
	"errors"
	"log"

	"github.com/daskol/nvim-bnf/bnf"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

var logger *Logger

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

	for line := from; line != to; line++ {
		if err := d.hightlightLine(v, buf, line); err != nil {
			logger.Warnf(
				"failed to hightlight line %d of %s: %s",
				line, buf, err,
			)
		}
	}
}

func (d *Document) hightlightLine(v *nvim.Nvim, buf nvim.Buffer, row int) error {
	defer func() {
		// TODO(@daskol): Test parser heavily!
		if ctx := recover(); ctx != nil {
			logger.Errorf("recovery: %s", ctx)
		}
	}()

	var ast *bnf.BNF
	var res int
	var err error

	if ast, err = bnf.Parse(d.Lines[row]); err != nil {
		return err
	}

	batch := v.NewBatch()
	batch.ClearBufferHighlight(buf, -1, row, row+1)

	bnf.Visit(ast.Rules[0], func(node bnf.Node) error {
		if node == nil {
			logger.Errorf("visiting nil node")
		}

		var grp string
		var begin, end int

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

	return batch.Execute()
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

// AttachToBuffer attaches plugin to buffer's updates. This method is temporary
// until it is supported in official Golang client.
func AttachToBuffer(v *nvim.Nvim, buf *nvim.Buffer) error {
	var result bool
	var args = []interface{}{
		buf,
		true,
		map[string]interface{}{},
	}

	if err := v.Invoke("nvim_buf_attach", &result, args...); err != nil {
		return err
	}

	if !result {
		return errors.New("nvim-bnf: result is false")
	}

	return nil
}

// DetachFromBuffer detaches plugin from buffer's updates. This method is
// temporary until it is supported in official Golang client.
func DetachFromBuffer(v *nvim.Nvim, buf *nvim.Buffer) error {
	var result bool
	var args = []interface{}{buf}

	if err := v.Invoke("nvim_buf_detach", &result, args...); err != nil {
		return err
	}

	if !result {
		return errors.New("nvim-bnf: result is false")
	}

	return nil
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
