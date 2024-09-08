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
func (r *Ruleset) HumanReadable() []byte {
	if len(*r) == 0 {
		return nil
	}
	var buf bytebuf.Chain
	r.hrHelper(&buf)
	return buf.Bytes()
}

// Internal human-readable helper.
func (r *Ruleset) hrHelper(buf *bytebuf.Chain) {
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString("<rules>\n")
	for _, rule := range *r {
		buf.WriteByte('\t')
		buf.WriteString(`<rule`)

		switch {
		case rule.callback != nil:
			r.attrB(buf, "callback", rule.src)
			r.hrArgs(buf, rule.arg)
		case rule.getter != nil:
			r.attrB(buf, "dst", rule.dst)
			r.attrB(buf, "getter", rule.src)
			r.hrArgs(buf, rule.arg)
		default:
			r.attrB(buf, "dst", rule.dst)
			if rule.static {
				r.attrB(buf, "src", rule.src)
				r.attrI(buf, "static", 1)
			} else {
				buf.WriteString(` src="`)
				r.hrVal(buf, rule.src, rule.subset)
				buf.WriteByte('"')
			}
			if len(rule.ins) > 0 {
				r.attrB(buf, "ins", rule.ins)
			}
		}

		if len(rule.mod) > 0 {
			buf.WriteByte('>')
		}
		if len(rule.mod) > 0 {
			buf.WriteByte('\n').WriteString("\t\t<mods>\n")
			for _, mod := range rule.mod {
				buf.WriteString("\t\t\t").WriteString(`<mod name="`).Write(mod.id).WriteByte('"')
				r.hrArgs(buf, mod.arg)
				buf.WriteString("/>\n")
			}
			buf.WriteString("\t\t</mods>\n")
		}

		if len(rule.mod) > 0 {
			buf.WriteByte('\t').WriteString("</rule>\n")
		} else {
			buf.WriteString("/>\n")
		}
	}
	buf.WriteString("</rules>\n")
}

// Human readable helper for value.
func (r *Ruleset) hrVal(buf *bytebuf.Chain, v []byte, set [][]byte) {
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
func (r *Ruleset) hrArgs(buf *bytebuf.Chain, args []*arg) {
	if len(args) > 0 {
		for j, a := range args {
			pfx := "arg"
			if a.static {
				pfx = "sarg"
			}
			buf.WriteByte(' ').WriteString(pfx).WriteInt(int64(j)).WriteString(`="`)
			r.hrVal(buf, a.val, a.subset)
			buf.WriteByte('"')
		}
	}
}

func (r *Ruleset) attrB(buf *bytebuf.Chain, key string, p []byte) {
	buf.WriteByte(' ').WriteString(key).WriteString(`="`).Write(bytes.ReplaceAll(p, hrQ, hrQR)).WriteByte('"')
}

func (r *Ruleset) attrS(buf *bytebuf.Chain, key, s string) {
	r.attrB(buf, key, byteconv.S2B(s))
}

func (r *Ruleset) attrI(buf *bytebuf.Chain, key string, i int) {
	buf.WriteByte(' ').WriteString(key).WriteString(`="`).WriteInt(int64(i)).WriteByte('"')
}
