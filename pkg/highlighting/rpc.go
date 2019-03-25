package highlighting

import (
	"errors"

	"github.com/neovim/go-client/nvim"
)

// NoOpts is a value which could be passed in NeoVim RPC.
var NoOpts = make(map[string]interface{})

// Chunk type describes part of virtual text.
type Chunk []string

// NewChunk creates new chunk of virtual text.
func NewChunk(text, hlGroup string) Chunk {
	return []string{text, hlGroup}
}

// SetVirtualText add virtual text to a buffer in batch mode.
func SetVirtualText(
	b *nvim.Batch, buf *nvim.Buffer, nsID int, line int, chunks []Chunk,
	opts map[string]interface{}, result *int,
) {
	var args = []interface{}{buf, nsID, line, &chunks, opts}
	b.Request("nvim_buf_set_virtual_text", result, args...)
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

	if err := v.Request("nvim_buf_attach", &result, args...); err != nil {
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

	if err := v.Request("nvim_buf_detach", &result, args...); err != nil {
		return err
	}

	if !result {
		return errors.New("nvim-bnf: result is false")
	}

	return nil
}
