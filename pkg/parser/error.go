package parser

import (
	"errors"
	"strconv"
)

var ErrEmptyRule = errors.New("bnf: rule is empty")
var ErrNoStatements = errors.New("bnf: there is no production statements")
var ErrNotImplemented = errors.New("bnf: not implemented")
var ErrUnexpectedChar = errors.New("bnf: unexpected character")

// Stringer is required here for human-readable error descriptions.
type Stringer interface {
	String() string
}

// Error represent parsing error and contains some context information.
type Error struct {
	err error
	pos int
}

func (e *Error) Error() string {
	return e.err.Error() + " at position " + strconv.Itoa(e.pos)
}

// DescError represents error which is occured during semantic parsing. It is
// based on Error but provides more human-readable representation with Stringer
// interface.
type DescError struct {
	Base Error
	desc string
}

func NewDescError(err error, pos int, desc string) *DescError {
	return &DescError{
		Base: Error{err: err, pos: pos},
		desc: desc,
	}
}

func (e *DescError) String() string {
	var pos = strconv.Itoa(e.Base.pos + 1)
	return "sem: " + e.desc + " is expected at position " + pos
}

func (e *DescError) Error() string {
	return e.Base.Error()
}
