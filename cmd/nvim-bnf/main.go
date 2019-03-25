// Tool nvim-bnf implement NeoVim MsgPack RPC and provides semantic syntax
// hightlighter for BNF.
package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

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
		doc := &Document{Lines: data}
		doc.Hightlight(v, *buf)
		DocIndex[*buf] = doc
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

func registerHandlers(p *plugin.Plugin) error {
	// Register event handlers during actual loading.
	if p.Nvim != nil {
		for _, event := range eventHandlers {
			p.Nvim.RegisterHandler(event.Name, event.Handler)
		}
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
}

type EventHandler struct {
	Name    string
	Handler interface{}
}

var flagGenManifest bool
var flagPluginHost string
var eventHandlers = []EventHandler{
	{"nvim_buf_changedtick_event", HandleBufChangedTickEvent},
	{"nvim_buf_detach_event", HandleBufDetachEvent},
	{"nvim_buf_lines_event", HandleBufLinesEvent},
}

func init() {
	flag.BoolVar(
		&flagGenManifest,
		"gen-manifest",
		false,
		"Trigger manifest generation instead of running of plugin")
	flag.StringVar(
		&flagPluginHost,
		"host",
		path.Base(os.Args[0]),
		"Set host name for manifest generator")
	flag.Parse()
}

func main() {
	var err error
	var nv *nvim.Nvim

	if logger, err = NewLogger(); err != nil {
		log.Fatalf("failed to instantiate logger: %s", err)
	}

	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("error occured during logger closing: %s", err)
		}
	}()

	defer func() {
		if nv != nil {
			nv.Close()
		}
	}()

	if flagGenManifest {
		var p = plugin.New(nil)

		if err = registerHandlers(p); err != nil {
			logger.Infof("failed to register handlers: %s", err)
			return
		}

		var manifest = p.Manifest(flagPluginHost)
		os.Stdout.Write(manifest)
	} else {
		var stdout = os.Stdout
		os.Stdout = os.Stderr

		if nv, err = nvim.New(os.Stdin, stdout, stdout, Log); err != nil {
			logger.Errorf("failed to create NeoVim client: %s", err)
			return
		}

		var p = plugin.New(nv)

		if err = registerHandlers(p); err != nil {
			logger.Errorf("failed to register plugin handlers: %s", err)
			return
		}

		if err = nv.Serve(); err != nil {
			logger.Errorf("failed to serve neovim plugin: %s", err)
			return
		}
	}
}
