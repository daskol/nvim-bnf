// Tool nvim-bnf implement NeoVim MsgPack RPC and provides semantic syntax
// hightlighter for BNF.
package main

import (
	"flag"
	"log"
	"os"
	"path"
)

var flagGenManifest bool
var flagPluginHost string

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
	if ptr, err := NewLogger(); err != nil {
		log.Fatalf("failed to instantiate logger: %s", err)
	} else {
		logger = ptr
	}

	defer func() {
		if err := logger.Close(); err != nil {
			log.Printf("error occured during logger closing: %s", err)
		}
	}()

	switch {
	case flagGenManifest:
		os.Stdout.Write(GenManifest(flagPluginHost))
	case !flagGenManifest:
		if err := RunPlugin(); err != nil {
			logger.Errorf("plugin was failed: %s", err)
		}
	}
}
