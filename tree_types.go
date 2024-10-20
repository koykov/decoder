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
	typeCond
	typeCondOK
	typeCondTrue
	typeCondFalse
	typeElse
	typeDiv
	typeSwitch
	typeCase
	typeDefault
)

// op represents a type of the operation in conditions and loops.
type op int

// Must be in sync with inspector.Op type.
const (
	opUnk op = iota
	opEq
	opNq
	opGt
	opGtq
	opLt
	opLtq
	opInc
	opDec
)

// Swap inverts itself.
func (o op) Swap() op {
	switch o {
	case opGt:
		return opLt
	case opGtq:
		return opLtq
	case opLt:
		return opGt
	case opLtq:
		return opGtq
	default:
		return o
	}
}

// String view of the operation.
func (o op) String() string {
	switch o {
	case opEq:
		return "=="
	case opNq:
		return "!="
	case opGt:
		return ">"
	case opGtq:
		return ">="
	case opLt:
		return "<"
	case opLtq:
		return "<="
	case opInc:
		return "++"
	case opDec:
		return "--"
	default:
		return "unk"
	}
}

// lc represents len/cap type.
type lc int

const (
	lcNone lc = iota
	lcLen
	lcCap
)

func (lc lc) String() string {
	switch lc {
	case lcLen:
		return "len"
	case lcCap:
		return "cap"
	case lcNone:
		fallthrough
	default:
		return ""
	}
}
