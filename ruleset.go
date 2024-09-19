package decoder

import (
	"bytes"

	"github.com/koykov/bytebuf"
	"github.com/koykov/byteconv"
)

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

// Ruleset represents list of rules.
type Ruleset []rule

// Rule object that describes one line in decoder's body.
type rule struct {
	typ rtype
	// Destination/source pair.
	dst, src, ins []byte
	// List of keys, that need to check sequentially in the source object.
	subset [][]byte
	// Getter callback, for sources like "dst = getFoo(var0, ...)"
	getter GetterFn
	// Callback for lines like "prepareObject(var.obj)"
	callback CallbackFn
	// Flag that indicates if source is a static value.
	static bool
	// List of modifier applied to source.
	mod []mod
	// List of arguments for getter or callback.
	arg []*arg
	// List of children nodes.
	child Ruleset

	// Loop stuff.
	loopKey       []byte
	loopVal       []byte
	loopSrc       []byte
	loopCnt       []byte
	loopCntInit   []byte
	loopCntStatic bool
	loopCntOp     Op
	loopCondOp    Op
	loopLim       []byte
	loopLimStatic bool
	loopBrkD      int
}

// Argument for getter/callback/modifier.
type arg struct {
	// Value argument.
	val []byte
	// List of keys, that need to check sequentially in the value object.
	subset [][]byte
	// Flag that indicates if value is a static value.
	static bool
}

var (
	hrQ  = []byte(`"`)
	hrQR = []byte(`&quot;`)
)

// HumanReadable builds human-readable view of the rules list.
func (rs *Ruleset) HumanReadable() []byte {
	if len(*rs) == 0 {
		return nil
	}
	var buf bytebuf.Chain
	rs.hrHelper(&buf, 0)
	return buf.Bytes()
}

// Internal human-readable helper.
func (rs *Ruleset) hrHelper(buf *bytebuf.Chain, depth int) {
	if depth == 0 {
		buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	}
	buf.WriteByteN('\t', depth).
		WriteString("<rules>\n")
	for _, r := range *rs {
		buf.WriteByteN('\t', depth+1)
		buf.WriteString(`<rule`)
		rs.attrI(buf, "type", int(r.typ))

		switch {
		case r.callback != nil:
			rs.attrB(buf, "callback", r.src)
			rs.hrArgs(buf, r.arg)
		case r.getter != nil:
			rs.attrB(buf, "dst", r.dst)
			rs.attrB(buf, "getter", r.src)
			rs.hrArgs(buf, r.arg)
		default:
			rs.attrB(buf, "dst", r.dst)
			if r.static {
				rs.attrB(buf, "src", r.src)
				rs.attrI(buf, "static", 1)
			} else if len(r.src) > 0 {
				buf.WriteString(` src="`)
				rs.hrVal(buf, r.src, r.subset)
				buf.WriteByte('"')
			}
			if len(r.ins) > 0 {
				rs.attrB(buf, "ins", r.ins)
			}
		}

		if len(r.mod) > 0 || len(r.child) > 0 {
			buf.WriteByte('>')
		}
		if len(r.mod) > 0 {
			buf.WriteByte('\n').
				WriteByteN('\t', depth+2).
				WriteString("<mods>\n")
			for _, mod := range r.mod {
				buf.WriteByteN('\t', depth+3).
					WriteString(`<mod name="`).Write(mod.id).WriteByte('"')
				rs.hrArgs(buf, mod.arg)
				buf.WriteString("/>\n")
			}
			buf.WriteByteN('\t', depth+2).
				WriteString("</mods>\n")
		}

		if len(r.mod) > 0 || len(r.child) > 0 {
			if len(r.child) > 0 {
				buf.WriteByte('\n')
				r.child.hrHelper(buf, depth+2)
			}
			buf.WriteByteN('\t', depth+1).
				WriteString("</rule>\n")
		} else {
			buf.WriteString("/>\n")
		}

	}
	buf.WriteByteN('\t', depth).WriteString("</rules>\n")
}

// Human readable helper for value.
func (rs *Ruleset) hrVal(buf *bytebuf.Chain, v []byte, set [][]byte) {
	buf.Write(v)
	if len(set) > 0 {
		buf.WriteString(".{")
		for i, s := range set {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.Write(s)
		}
		buf.WriteByte('}')
	}
}

// Human-readable helper for args list.
func (rs *Ruleset) hrArgs(buf *bytebuf.Chain, args []*arg) {
	if len(args) > 0 {
		for j, a := range args {
			pfx := "arg"
			if a.static {
				pfx = "sarg"
			}
			buf.WriteByte(' ').
				WriteString(pfx).
				WriteInt(int64(j)).
				WriteString(`="`)
			rs.hrVal(buf, a.val, a.subset)
			buf.WriteByte('"')
		}
	}
}

func (rs *Ruleset) attrB(buf *bytebuf.Chain, key string, p []byte) {
	if len(p) == 0 {
		return
	}
	buf.WriteByte(' ').
		WriteString(key).
		WriteString(`="`).
		Write(bytes.ReplaceAll(p, hrQ, hrQR)).
		WriteByte('"')
}

func (rs *Ruleset) attrS(buf *bytebuf.Chain, key, s string) {
	rs.attrB(buf, key, byteconv.S2B(s))
}

func (rs *Ruleset) attrI(buf *bytebuf.Chain, key string, i int) {
	buf.WriteByte(' ').
		WriteString(key).
		WriteString(`="`).
		WriteInt(int64(i)).
		WriteByte('"')
}
