package decoder

import "bytes"

// List of rules.
type Ruleset []rule

// Rule object that describes one line in decoder's body.
type rule struct {
	// Destination/source pair.
	dst, src, ins []byte
	// List of keys, that need to check sequentially in the source object.
	subset [][]byte
	// Getter callback, for sources like "dst = getFoo(var0, ...)"
	getter *GetterFn
	// Callback for lines like "prepareObject(var.obj)"
	callback *CallbackFn
	// Flag that indicates if source is a static value.
	static bool
	// List of modifier applied to source.
	mod []mod
	// List of arguments for getter or callback.
	arg []*arg
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

// Build human readable view of the rules list.
func (r *Ruleset) HumanReadable() []byte {
	if len(*r) == 0 {
		return nil
	}
	var buf bytes.Buffer
	r.hrHelper(&buf)
	return buf.Bytes()
}

// Internal human readable helper.
func (r *Ruleset) hrHelper(buf *bytes.Buffer) {
	for _, rule := range *r {
		if rule.callback == nil {
			buf.WriteString("dst: ")
			buf.Write(rule.dst)
			buf.WriteString(" <- src: ")
			if rule.static {
				buf.WriteByte('"')
				buf.Write(rule.src)
				buf.WriteByte('"')
			} else {
				r.hrVal(buf, rule.src, rule.subset)
			}
			if len(rule.ins) > 0 {
				buf.WriteString(" as ")
				buf.Write(rule.ins)
			}
		} else {
			buf.WriteString("cb: ")
			buf.Write(rule.src)
		}

		if rule.getter != nil || rule.callback != nil {
			buf.WriteByte('(')
			for i, a := range rule.arg {
				if i > 0 {
					buf.WriteByte(',')
					buf.WriteByte(' ')
				}
				if a.static {
					buf.WriteByte('"')
					buf.Write(a.val)
					buf.WriteByte('"')
				} else {
					r.hrVal(buf, a.val, a.subset)
				}
			}
			buf.WriteByte(')')
		}

		if len(rule.mod) > 0 {
			buf.WriteString(" mod")
			for i, mod := range rule.mod {
				if i > 0 {
					buf.WriteByte(',')
				}
				buf.WriteByte(' ')
				buf.Write(mod.id)
				if len(mod.arg) > 0 {
					buf.WriteByte('(')
					for j, a := range mod.arg {
						if j > 0 {
							buf.WriteByte(',')
							buf.WriteByte(' ')
						}
						if a.static {
							buf.WriteByte('"')
							buf.Write(a.val)
							buf.WriteByte('"')
						} else {
							buf.Write(a.val)
						}
					}
					buf.WriteByte(')')
				}
			}
		}

		buf.WriteByte('\n')
	}
}

// Human readable helper for value.
func (r *Ruleset) hrVal(buf *bytes.Buffer, v []byte, set [][]byte) {
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
