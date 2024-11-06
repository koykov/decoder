package decoder

import (
	"bytes"

	"github.com/koykov/bytebuf"
	"github.com/koykov/byteconv"
)

type Ruleset []node

// Tree represents list of nodes.
type Tree struct {
	nodes Ruleset
	hsum  uint64
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

// Ruleset returns list of root nodes as old ruleset.
// Implements to support old types.
func (t *Tree) Ruleset() Ruleset {
	return t.nodes
}

// HumanReadable builds human-readable view of the nodes list.
func (t *Tree) HumanReadable() []byte {
	if len(t.nodes) == 0 {
		return nil
	}
	var buf bytebuf.Chain
	t.hrHelper(&buf, t.nodes, 0)
	return buf.Bytes()
}

// Internal human-readable helper.
func (t *Tree) hrHelper(buf *bytebuf.Chain, nodes []node, depth int) {
	if depth == 0 {
		buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	}
	buf.WriteByteN('\t', depth).
		WriteString("<nodes>\n")
	for _, n := range nodes {
		buf.WriteByteN('\t', depth+1)
		buf.WriteString(`<node`)
		t.attrIF(buf, "type", int(n.typ), depth == 0)

		switch {
		case n.callback != nil:
			t.attrB(buf, "callback", n.src)
			t.hrArgs(buf, n.arg)
		case n.getter != nil:
			t.attrB(buf, "dst", n.dst)
			t.attrB(buf, "getter", n.src)
			t.hrArgs(buf, n.arg)
		default:
			t.attrB(buf, "dst", n.dst)
			if n.static {
				t.attrB(buf, "src", n.src)
				t.attrI(buf, "static", 1)
			} else if len(n.src) > 0 {
				buf.WriteString(` src="`)
				t.hrVal(buf, n.src, n.subset)
				buf.WriteByte('"')
			}
			if len(n.ins) > 0 {
				t.attrB(buf, "ins", n.ins)
			}
		}

		if n.typ == typeCond {
			if len(n.condL) > 0 {
				t.attrB(buf, "left", n.condL)
			}
			if n.condOp != 0 {
				t.attrS(buf, "op", n.condOp.String())
			}
			if len(n.condR) > 0 {
				t.attrB(buf, "right", n.condR)
			}
			if len(n.condHlp) > 0 {
				t.attrB(buf, "helper", n.condHlp)
				if n.condLC > lcNone {
					t.attrS(buf, "lc", n.condLC.String())
				}
				if len(n.condHlpArg) > 0 {
					for j, a := range n.condHlpArg {
						pfx := "arg"
						if a.static {
							pfx = "sarg"
						}
						buf.WriteByte(' ').
							WriteString(pfx).
							WriteInt(int64(j)).
							WriteString(`="`).
							Write(a.val).
							WriteByte('"')
					}
				}
			}
		}

		if n.typ == typeCondOK {
			t.attrB(buf, "var", n.condOKL)
			t.attrB(buf, "varOK", n.condOKR)

			if len(n.condHlp) > 0 {
				t.attrB(buf, "helper", n.condHlp)
				if len(n.condHlpArg) > 0 {
					for j, a := range n.condHlpArg {
						pfx := "arg"
						if a.static {
							pfx = "sarg"
						}
						buf.WriteByte(' ').
							WriteString(pfx).
							WriteInt(int64(j)).
							WriteString(`="`).
							Write(a.val).
							WriteByte('"')
					}
				}
			}

			if len(n.condL) > 0 {
				t.attrB(buf, "left", n.condL)
			}
			if n.condOp != 0 {
				t.attrS(buf, "op", n.condOp.String())
			}
			if len(n.condR) > 0 {
				t.attrB(buf, "right", n.condR)
			}
		}

		if n.typ == typeCase {
			t.attrB(buf, "left", n.caseL)
			t.attrBl(buf, "leftStatic", n.caseStaticL)
			t.attrS(buf, "op", n.caseOp.String())
			t.attrB(buf, "right", n.caseR)
			t.attrBl(buf, "rightStatic", n.caseStaticR)
			t.attrB(buf, "hlp", n.caseHlp)
			if len(n.caseHlpArg) > 0 {
				for j, a := range n.caseHlpArg {
					pfx := "arg"
					if a.static {
						pfx = "sarg"
					}
					buf.WriteByte(' ').
						WriteString(pfx).
						WriteInt(int64(j)).
						WriteString(`="`).
						Write(a.val).
						WriteByte('"')
				}
			}
		}

		if n.typ == typeLoopCount || n.typ == typeLoopRange {
			t.attrB(buf, "key", n.loopKey)
			t.attrB(buf, "val", n.loopVal)
			t.attrB(buf, "src", n.loopSrc)
			t.attrB(buf, "counter", n.loopCnt)
			t.attrS(buf, "cond", n.loopCondOp.String())
			t.attrB(buf, "limit", n.loopLim)
			t.attrS(buf, "op", n.loopCntOp.String())
		}
		t.attrI(buf, "brkD", n.loopBrkD)

		if len(n.mod) > 0 || len(n.child) > 0 {
			buf.WriteByte('>')
		}
		if len(n.mod) > 0 {
			buf.WriteByte('\n').
				WriteByteN('\t', depth+2).
				WriteString("<mods>\n")
			for _, mod := range n.mod {
				buf.WriteByteN('\t', depth+3).
					WriteString(`<mod name="`).Write(mod.id).WriteByte('"')
				t.hrArgs(buf, mod.arg)
				buf.WriteString("/>\n")
			}
			buf.WriteByteN('\t', depth+2).
				WriteString("</mods>\n")
		}

		if len(n.mod) > 0 || len(n.child) > 0 {
			if len(n.child) > 0 {
				buf.WriteByte('\n')
				t.hrHelper(buf, n.child, depth+2)
			}
			buf.WriteByteN('\t', depth+1).
				WriteString("</node>\n")
		} else {
			buf.WriteString("/>\n")
		}

	}
	buf.WriteByteN('\t', depth).WriteString("</nodes>\n")
}

// Human readable helper for value.
func (t *Tree) hrVal(buf *bytebuf.Chain, v []byte, set [][]byte) {
	if bytes.IndexByte(v, '"') != -1 {
		v = bytes.ReplaceAll(v, hrQ, hrQR)
	}
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
func (t *Tree) hrArgs(buf *bytebuf.Chain, args []*arg) {
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
			t.hrVal(buf, a.val, a.subset)
			buf.WriteByte('"')
		}
	}
}

func (t *Tree) attrB(buf *bytebuf.Chain, key string, p []byte) {
	if len(p) == 0 {
		return
	}
	buf.WriteByte(' ').
		WriteString(key).
		WriteString(`="`).
		Write(bytes.ReplaceAll(p, hrQ, hrQR)).
		WriteByte('"')
}

func (t *Tree) attrS(buf *bytebuf.Chain, key, s string) {
	t.attrB(buf, key, byteconv.S2B(s))
}

func (t *Tree) attrBl(buf *bytebuf.Chain, key string, b bool) {
	v := "1"
	if !b {
		v = ""
	}
	t.attrB(buf, key, byteconv.S2B(v))
}

func (t *Tree) attrI(buf *bytebuf.Chain, key string, i int) {
	t.attrIF(buf, key, i, false)
}

func (t *Tree) attrIF(buf *bytebuf.Chain, key string, i int, force bool) {
	if i == 0 && !force {
		return
	}
	buf.WriteByte(' ').
		WriteString(key).
		WriteString(`="`).
		WriteInt(int64(i)).
		WriteByte('"')
}
