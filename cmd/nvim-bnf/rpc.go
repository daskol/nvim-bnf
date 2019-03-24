package main

import (
	"errors"

	"github.com/daskol/go-client/nvim"
)

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
