package decoder

// rtype of the node.
type rtype int

const (
	typeOperator rtype = iota
	typeLoopRange
	typeLoopCount
	typeLBreak
	typeBreak
	typeContinue
)
