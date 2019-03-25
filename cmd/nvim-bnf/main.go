// Tool nvim-bnf implement NeoVim MsgPack RPC and provides semantic syntax
// hightlighter for BNF.
package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/daskol/nvim-bnf/pkg/highlighting"
	"github.com/daskol/nvim-bnf/pkg/logging"
)

var flagGenManifest bool
var flagPluginHost string
var logger = logging.Get()

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
	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("error occured during logger closing: %s", err)
		}
	}()

	switch {
	case flagGenManifest:
		os.Stdout.Write(highlighting.GenManifest(flagPluginHost))
	case !flagGenManifest:
		if err := highlighting.RunPlugin(); err != nil {
			logger.Errorf("plugin was failed: %s", err)
		}
	}
}
