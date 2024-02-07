package decoder

import "errors"

var (
	ErrDecoderNotFound = errors.New("decoder not found")
	ErrEmptyNode       = errors.New("provided node is empty")
	ErrModNoArgs       = errors.New("empty arguments list")
	ErrModPoorArgs     = errors.New("arguments list in modifier is too small")
	ErrCbPoorArgs      = errors.New("arguments list in callback is too small")
	ErrGetterPoorArgs  = errors.New("arguments list in getter callback is too small")

	ErrUnknownPool = errors.New("unknown pool")

	_ = ErrCbPoorArgs
)
