package decoder

import "errors"

var (
	ErrDecoderNotFound = errors.New("decoder not found")
	ErrEmptyNode       = errors.New("provided node is empty")
	ErrCbPoorArgs      = errors.New("arguments list in callback is too small")
	ErrGetterPoorArgs  = errors.New("arguments list in getter callback is too small")
)
