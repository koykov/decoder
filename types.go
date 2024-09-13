package decoder

// Op represents a type of the operation in conditions and loops.
type Op int

const (
	OpUnk Op = iota
	OpEq
	OpNq
	OpGt
	OpGtq
	OpLt
	OpLtq
	OpInc
	OpDec
)

func (o Op) String() string {
	switch o {
	case OpEq:
		return "=="
	case OpNq:
		return "!="
	case OpGt:
		return ">"
	case OpGtq:
		return ">="
	case OpLt:
		return "<"
	case OpLtq:
		return "<="
	case OpInc:
		return "++"
	case OpDec:
		return "--"
	case OpUnk:
		fallthrough
	default:
		return "unk"
	}
}
