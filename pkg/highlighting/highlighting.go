// Package highlighting contains provides NeoVim RPC client for syntax
// highlighting.
package highlighting

import (
	"os"

	"github.com/daskol/nvim-bnf/pkg/logging"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

var logger = logging.Get()

// GenManifest generates a remote plugin manifest. It is parametrized with
// plugin host name. In this particular case host name is name of plugin
// binary.
func GenManifest(host string) []byte {
	hl := new(Highlighter)
	hl.plugin = plugin.New(nil)
	hl.registerVimLExtHandlers()
	return hl.plugin.Manifest(host)
}

// RunPlugin initializes NeoVim RPC client, registers local procedures and runs
// plugin over stdint/stdout pair.
func RunPlugin() error {
	var err error
	var hl Highlighter
	var log = logging.Log
	var stdout = os.Stdout
	os.Stdout = os.Stderr

	if hl.nvim, err = nvim.New(os.Stdin, stdout, stdout, log); err != nil {
		logger.Infof("failed to create neovim client")
		return err
	}

	hl.plugin = plugin.New(hl.nvim)

	if err = hl.registerHandlers(); err != nil {
		logger.Errorf("failed to register plugin handlers")
		return err
	}

	return hl.Serve()
}

// Highlighter is an implementation of semantic hightlighting for BNF. It
// manages all RPC request and response between NeoVim instance and BNF parser.
type Highlighter struct {
	nvim   *nvim.Nvim
	plugin *plugin.Plugin
}

func (h *Highlighter) HandleBufReadEvent(buf nvim.Buffer, filename string) {
	logger.Debugf("HandleBufReadEvent(%s)", filename)

	if err := AttachToBuffer(h.nvim, &buf); err != nil {
		logger.Errorf("failed to attach to buffer: %s", err)
		return
	}

	logger.Infof("buffer %s was attached to plugin", buf)
}

func (h *Highlighter) HandleBufLinesEvent(
	buf *nvim.Buffer, changedTick int, firstLine, lastLine int,
	data [][]byte, more bool,
) {
	logger.Debugf(
		"HandleBufLinesEvent(%s, %d, %d, %d, [[...]], %t)",
		buf, changedTick, firstLine, lastLine, more,
	)

	if lastLine == -1 {
		doc := &Document{Lines: data}
		doc.Hightlight(h.nvim, *buf)
		DocIndex[*buf] = doc
	} else {
		var doc, ok = DocIndex[*buf]

		if !ok {
			logger.Warnf("unknown buffer: %s", buf)
			return
		}

		var from, to = doc.Update(data, firstLine, lastLine)
		doc.HightlightHunk(h.nvim, *buf, from, to)
	}
}

func (p *Highlighter) HandleBufDetachEvent(buf *nvim.Buffer) {
	logger.Debugf("HandleBufDetachEvent(%s)", buf)

	if err := DetachFromBuffer(p.nvim, buf); err != nil {
		logger.Errorf("failed to detac buffer: %s", err)
		return
	}

	logger.Infof("buffer %d was detached from plugin", buf)
}

func (h *Highlighter) HandleBufChangedTickEvent(
	buf nvim.Buffer, changedTick int,
) {
	logger.Debugf("HandleBufChangedTickEvent(%s, %d)", buf, changedTick)
}

func (h *Highlighter) HandleNcm2OnWarmup(args []interface{}) {
	if len(args) != 1 {
		logger.Errorf("HandleNcm2OnWarmup(): too few arguments")
		return
	}

	var ctx, ok = args[0].(map[string]interface{})

	if !ok {
		logger.Errorf("HandleNcm2OnWarmup(): wrong argument type")
		return
	}

	logger.Debugf("HandleNcm2OnWarmup(%s)", ctx)
}

func (h *Highlighter) HandleNcm2OnComplete(args []interface{}) {
	if len(args) != 1 {
		logger.Errorf("HandleNcm2OnComplete(): too few arguments")
		return
	}

	var ctx, ok = args[0].(map[string]interface{})

	if !ok {
		logger.Errorf("HandleNcm2OnComplete(): wrong argument type")
		return
	}

	h.handleNCM2OnComplete(ctx)
}

func (h *Highlighter) handleNCM2OnComplete(ctx map[string]interface{}) {
	logger.Debugf("HandleNcm2OnComplete(%s)", ctx)
	var startccol = ctx["startccol"].(int64)
	var matches = []map[string]interface{}{}

	err := h.nvim.Call("ncm2#complete", nil, ctx, startccol, matches)

	if err != nil {
		logger.Errorf("failed call ncm2#complete: %s", err)
		return
	}
}

func (h *Highlighter) Serve() error {
	return h.nvim.Serve()
}

func (h *Highlighter) registerHandlers() error {
	h.registerVimLExtHandlers()
	return h.registerEventHandlers()
}

func (h *Highlighter) registerAutocmdHandlers() {
	// Register autocommands.
	for _, event := range []string{"BufRead", "BufNewFile"} {
		var opts = &plugin.AutocmdOptions{
			Event:   event,
			Group:   "nvim-bnf",
			Pattern: "*.bnf",
			Eval:    `expand("<afile>")`,
		}
		h.plugin.HandleAutocmd(opts, h.HandleBufReadEvent)
	}
}

func (h *Highlighter) registerEventHandlers() error {
	var eventHandlers = []struct {
		name    string
		handler interface{}
	}{
		{"nvim_buf_changedtick_event", h.HandleBufChangedTickEvent},
		{"nvim_buf_detach_event", h.HandleBufDetachEvent},
		{"nvim_buf_lines_event", h.HandleBufLinesEvent},
	}

	// Register event handlers during loading in operational mode.
	for _, event := range eventHandlers {
		var err = h.nvim.RegisterHandler(event.name, event.handler)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Highlighter) registerFunctionHandlers() {
	type FuncOpts = plugin.FunctionOptions
	var functions = []struct {
		name    string
		handler interface{}
	}{
		{"BNFNcm2OnWarmup", h.HandleNcm2OnWarmup},
		{"BNFNcm2OnComplete", h.HandleNcm2OnComplete},
	}

	// Register event handlers during loading in operational mode.
	for _, proc := range functions {
		h.plugin.HandleFunction(&FuncOpts{Name: proc.name}, proc.handler)
	}
}

func (h *Highlighter) registerVimLExtHandlers() {
	h.registerAutocmdHandlers()
	h.registerFunctionHandlers()
}
