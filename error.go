package jsondecoder

import "errors"

var (
	ErrDecoderNotFound = errors.New("decoder not found")
	ErrCbPoorArgs      = errors.New("arguments list in callback is too small")
	ErrGetterPoorArgs  = errors.New("arguments list in getter callback is too small")
)
