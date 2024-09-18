package decoder

import "errors"

var (
	ErrDecoderNotFound = errors.New("decoder not found")
	ErrEmptyNode       = errors.New("provided node is empty")
	ErrModNoArgs       = errors.New("empty arguments list")
	ErrModPoorArgs     = errors.New("arguments list in modifier is too small")
	ErrCbPoorArgs      = errors.New("arguments list in callback is too small")
	ErrGetterPoorArgs  = errors.New("arguments list in getter callback is too small")

	ErrUnbalancedCtl = errors.New("unbalanced control structures found")

	ErrWrongLoopLim  = errors.New("wrong count loop limit argument")
	ErrWrongLoopCond = errors.New("wrong loop condition operation")
	ErrWrongLoopOp   = errors.New("wrong loop operation")
	ErrBreakLoop     = errors.New("break loop")
	ErrLBreakLoop    = errors.New("lazybreak loop")
	ErrContLoop      = errors.New("continue loop")

	ErrUnknownPool = errors.New("unknown pool")

	_ = ErrCbPoorArgs
)
