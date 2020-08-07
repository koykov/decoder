package jsondecoder

import "errors"

var (
	ErrDecoderNotFound = errors.New("decoder not found")
	ErrGetterPoorArgs  = errors.New("arguments list in getter callback is too small")
)
