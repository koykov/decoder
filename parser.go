package decoder

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/koykov/bytealg"
	"github.com/koykov/byteconv"
)

type Parser struct {
	// Decoder body to parse.
	body []byte

	// Counters (depths) of conditions, loops and switches.
	cc, cl, cs int
}

var (
	// Byte constants.
	nl      = []byte("\n")
	vline   = []byte("|")
	space   = []byte(" ")
	comma   = []byte(",")
	uscore  = []byte("_")
	dot     = []byte(".")
	empty   = []byte("")
	qbO     = []byte("[")
	qbC     = []byte("]")
	noFmt   = []byte(" \t\n\r")
	quotes  = []byte("\"'`")
	comment = []byte("//")

	// Operation constants.
	opEq  = []byte("==")
	opNq  = []byte("!=")
	opGt  = []byte(">")
	opGtq = []byte(">=")
	opLt  = []byte("<")
	opLtq = []byte("<=")
	opInc = []byte("++")
	opDec = []byte("--")

	// Regexp to parse expressions.
	reAssignV2CAs  = regexp.MustCompile(`((?:context|ctx)\.[\w\d\\.\[\]]+)\s*=\s*(.*) as ([:\w]*)`)
	reAssignV2CDot = regexp.MustCompile(`((?:context|ctx)\.[\w\d\\.\[\]]+)\s*=\s*(.*).\(([:\w]*)\)`)
	reAssignV2C    = regexp.MustCompile(`((?:context|ctx)\.[\w\d\\.\[\]]+)\s*=\s*(.*)`)

	reAssignV2V = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*(.*)`)
	reAssignF2V = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*([^(|]+)\(([^)]*)\)`)
	reFunction  = regexp.MustCompile(`([^(]+)\(([^)]*)\)`)
	reMod       = regexp.MustCompile(`([^(]+)\(*([^)]*)\)*`)
	reSet       = regexp.MustCompile(`(.*)\.{([^}]+)}`)

	reLoop      = regexp.MustCompile(`for .*`)
	reLoopRange = regexp.MustCompile(`for ([^:]+)\s*:*=\s*range\s*([^\s]*)\s*\{` + "")
	reLoopCount = regexp.MustCompile(`for (\w*)\s*:*=\s*(\w+)\s*;\s*\w+\s*(<|<=|>|>=|!=)+\s*([^;]+)\s*;\s*\w*(--|\+\+)+\s*\{`)

	// Suppress go vet warning.
	_ = ParseFile
)

// Parse parses the decoder rules.
func Parse(src []byte) (Ruleset, error) {
	p := &Parser{body: src}
	t := newTarget(p)
	ruleset, _, err := p.parse1(nil, 0, t)
	return ruleset, err
}

// ParseFile parses the file.
func ParseFile(fileName string) (rules Ruleset, err error) {
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	var raw []byte
	raw, err = os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %s", fileName)
	}
	return Parse(raw)
}

// func (p *Parser) parse() (ruleset Ruleset, err error) {
// 	// Split body to separate lines.
// 	// Each line contains only one expression.
// 	lines := bytes.Split(p.body, nl)
// 	ruleset = make(Ruleset, 0, len(lines))
// 	for i, line := range lines {
// 		if len(line) == 0 || bytes.HasPrefix(line[:2], comment) {
// 			continue
// 		}
// 		line = bytealg.Trim(line, noFmt)
// 		if len(line) == 0 || line[0] == '#' {
// 			// Ignore comments.
// 			continue
// 		}
// 		r := rule{typ: typeOperator}
// 		switch {
// 		case reAssignV2C.Match(line):
// 			// Var-to-ctx expression caught.
// 			if m := reAssignV2CAs.FindSubmatch(line); m != nil {
// 				r.dst = m[1]
// 				r.src = m[2]
// 				r.ins = m[3]
// 			} else if m = reAssignV2CDot.FindSubmatch(line); m != nil {
// 				r.dst = m[1]
// 				r.src = m[2]
// 				r.ins = m[3]
// 			} else if m = reAssignV2C.FindSubmatch(line); m != nil {
// 				r.dst = m[1]
// 				r.src = m[2]
// 			}
// 			// Check static/variable.
// 			if r.static = isStatic(r.src); r.static {
// 				r.src = bytealg.Trim(r.src, quotes)
// 			} else {
// 				r.src, r.mod = extractMods(r.src)
// 				r.src, r.subset = extractSet(r.src)
// 			}
// 			ruleset = append(ruleset, r)
// 			continue
// 		case reAssignV2V.Match(line):
// 			// Var-to-var expression caught.
// 			if m := reAssignF2V.FindSubmatch(line); m != nil {
// 				// Func-to-var expression caught.
// 				r.dst = replaceQB(m[1])
// 				r.src = replaceQB(m[2])
// 				// Parse getter callback.
// 				fn := GetGetterFn(byteconv.B2S(m[2]))
// 				if fn != nil {
// 					r.getter = fn
// 					r.arg = extractArgs(m[3])
// 				} else {
// 					// Getter func not found, so try to fallback to mod func.
// 					m = reAssignV2V.FindSubmatch(line)
// 					r.src, r.mod = extractMods(m[2])
// 					r.src, r.subset = extractSet(r.src)
// 				}
// 				if r.getter == nil && len(r.mod) == 0 {
// 					err = fmt.Errorf("unknown getter nor modifier function '%s' at line %d", m[2], i)
// 					break
// 				}
// 			} else if m = reAssignV2V.FindSubmatch(line); m != nil {
// 				// Var-to-var ...
// 				r.dst = replaceQB(m[1])
// 				if r.static = isStatic(m[2]); r.static {
// 					r.src = bytealg.Trim(m[2], quotes)
// 				} else {
// 					r.src, r.mod = extractMods(m[2])
// 					r.src, r.subset = extractSet(r.src)
// 				}
// 			}
// 			ruleset = append(ruleset, r)
// 			continue
// 		case reFunction.Match(line):
// 			m := reFunction.FindSubmatch(line)
// 			// Function expression caught.
// 			r.src = m[1]
// 			// Parse callback.
// 			fn := GetCallbackFn(byteconv.B2S(m[1]))
// 			if fn == nil {
// 				err = fmt.Errorf("unknown callback function '%s' at line %d", m[1], i)
// 				break
// 			}
// 			r.callback = fn
// 			r.arg = extractArgs(m[2])
//
// 			ruleset = append(ruleset, r)
// 			continue
// 		case reLoop.Match(line):
// 			if m := reLoopRange.FindSubmatch(line); m != nil {
// 				r.typ = typeLoopRange
// 				// todo implement me
// 			} else if m := reLoopCount.FindSubmatch(line); m != nil {
// 				r.typ = typeLoopCount
// 				// todo implement me
// 			} else {
// 				return ruleset, fmt.Errorf("couldn't parse loop control structure '%s' at offset %d", line, i)
// 			}
// 			t := newTarget(p)
// 			_ = t
// 			p.cl++
//
// 			if r.child, err = p.parse(); err != nil {
// 				return ruleset, err
// 			}
// 			ruleset = append(ruleset, r)
// 			continue
// 		case true:
// 			// todo check loop end
// 		}
// 		// Report unparsed error.
// 		err = fmt.Errorf("unknown rule '%s' a line %d", line, i)
// 		break
// 	}
// 	return
// }

func (p *Parser) parse1(dst Ruleset, offset int, t *target) (Ruleset, int, error) {
	var (
		ctl     []byte
		eol, up bool
		err     error
	)
	for !eol {
		ctl, offset, eol = p.nextCtl(offset)
		if len(ctl) == 0 {
			continue
		}
		r := rule{typ: typeOperator}
		if dst, up, err = p.processCtl(dst, &r, ctl, offset); err != nil {
			return dst, offset, err
		}
		offset += len(ctl)
		if up {
			break
		}

	}
	return dst, offset, nil
}

func (p *Parser) nextCtl(offset int) ([]byte, int, bool) {
	var eol bool
	if offset, eol = p.skipFmt(offset); eol {
		return nil, offset, true
	}
	i := bytes.IndexAny(p.body[offset:], "\n\r")
	if i == -1 {
		return p.body[offset:], offset, false
	}
	ctl := p.body[offset : offset+i]
	if j := bytes.IndexByte(ctl, '{'); j > 0 && ctl[j-1] != '.' {
		return p.body[offset : offset+j+1], offset, false
	}
	return p.body[offset : offset+i], offset, false
}

func (p *Parser) processCtl(dst Ruleset, root *rule, ctl []byte, offset int) (Ruleset, bool, error) {
	var err error
	switch {
	case ctl[0] == '#' || bytes.HasPrefix(ctl, comment):
		return dst, false, nil
	case reLoop.Match(ctl):
		if m := reLoopRange.FindSubmatch(ctl); m != nil {
			root.typ = typeLoopRange
			if bytes.Contains(m[1], comma) {
				kv := bytes.Split(m[1], comma)
				root.loopKey = bytealg.Trim(kv[0], space)
				if bytes.Equal(root.loopKey, uscore) {
					root.loopKey = nil
				}
				root.loopVal = bytealg.Trim(kv[1], space)
			} else {
				root.loopKey = bytealg.Trim(m[1], space)
			}
			root.loopSrc = m[2]
		} else if m := reLoopCount.FindSubmatch(ctl); m != nil {
			root.typ = typeLoopCount
			root.loopCnt = m[1]
			root.loopCntInit = m[2]
			root.loopCntStatic = isStatic(m[2])
			root.loopCondOp = p.parseOp(m[3])
			root.loopLim = m[4]
			root.loopLimStatic = isStatic(m[4])
			root.loopCntOp = p.parseOp(m[5])
		} else {
			return dst, false, fmt.Errorf("couldn't parse loop control structure '%s' at offset %d", string(ctl), offset)
		}
		t := newTarget(p)
		_ = t
		p.cl++

		offset += len(ctl)
		if root.child, offset, err = p.parse1(root.child, offset, t); err != nil {
			return dst, false, err
		}
		dst = append(dst, *root)
	case ctl[0] == '}':
		// todo check target and exit from current branch (loop/switch/condition)
		return dst, true, err
	case reAssignV2C.Match(ctl):
		// Var-to-ctx expression caught.
		if m := reAssignV2CAs.FindSubmatch(ctl); m != nil {
			root.dst = m[1]
			root.src = m[2]
			root.ins = m[3]
		} else if m = reAssignV2CDot.FindSubmatch(ctl); m != nil {
			root.dst = m[1]
			root.src = m[2]
			root.ins = m[3]
		} else if m = reAssignV2C.FindSubmatch(ctl); m != nil {
			root.dst = m[1]
			root.src = m[2]
		}
		// Check static/variable.
		if root.static = isStatic(root.src); root.static {
			root.src = bytealg.Trim(root.src, quotes)
		} else {
			root.src, root.mod = extractMods(root.src)
			root.src, root.subset = extractSet(root.src)
		}
		dst = append(dst, *root)
	case reAssignV2V.Match(ctl):
		// Var-to-var expression caught.
		if m := reAssignF2V.FindSubmatch(ctl); m != nil {
			// Func-to-var expression caught.
			root.dst = replaceQB(m[1])
			root.src = replaceQB(m[2])
			// Parse getter callback.
			fn := GetGetterFn(byteconv.B2S(m[2]))
			if fn != nil {
				root.getter = fn
				root.arg = extractArgs(m[3])
			} else {
				// Getter func not found, so try to fallback to mod func.
				m = reAssignV2V.FindSubmatch(ctl)
				root.src, root.mod = extractMods(m[2])
				root.src, root.subset = extractSet(root.src)
			}
			if root.getter == nil && len(root.mod) == 0 {
				err = fmt.Errorf("unknown getter nor modifier function '%s' at offset %d", m[2], offset)
				break
			}
		} else if m = reAssignV2V.FindSubmatch(ctl); m != nil {
			// Var-to-var ...
			root.dst = replaceQB(m[1])
			if root.static = isStatic(m[2]); root.static {
				root.src = bytealg.Trim(m[2], quotes)
			} else {
				root.src, root.mod = extractMods(m[2])
				root.src, root.subset = extractSet(root.src)
			}
		}
		dst = append(dst, *root)
	case reFunction.Match(ctl):
		m := reFunction.FindSubmatch(ctl)
		// Function expression caught.
		root.src = m[1]
		// Parse callback.
		fn := GetCallbackFn(byteconv.B2S(m[1]))
		if fn == nil {
			err = fmt.Errorf("unknown callback function '%s' at offset %d", m[1], offset)
			break
		}
		root.callback = fn
		root.arg = extractArgs(m[2])

		dst = append(dst, *root)
	default:
		return dst, false, fmt.Errorf("unknown rule '%s' at position %d", string(ctl), offset)
	}
	return dst, false, nil
}

func (p *Parser) skipFmt(offset int) (int, bool) {
	n := len(p.body)
	for i := offset; i < n; i++ {
		c := p.body[i]
		if c != '\n' && c != '\r' && c != '\t' && c != ' ' {
			return i, i == n-1
		}
	}
	return n - 1, true
}

func (p *Parser) parseOp(src []byte) Op {
	var op Op
	switch {
	case bytes.Equal(src, opEq):
		op = OpEq
	case bytes.Equal(src, opNq):
		op = OpNq
	case bytes.Equal(src, opGt):
		op = OpGt
	case bytes.Equal(src, opGtq):
		op = OpGtq
	case bytes.Equal(src, opLt):
		op = OpLt
	case bytes.Equal(src, opLtq):
		op = OpLtq
	case bytes.Equal(src, opInc):
		op = OpInc
	case bytes.Equal(src, opDec):
		op = OpDec
	default:
		op = OpUnk
	}
	return op
}

// Split expression to variable and mods list.
func extractMods(p []byte) ([]byte, []mod) {
	hasVline := bytes.Contains(p, vline)
	hasSet := reSet.Match(p)
	modNoVar := reFunction.Match(p) && !hasVline
	if (hasVline && !hasSet) || modNoVar {
		mods := make([]mod, 0)
		chunks := bytes.Split(p, vline)
		var idx = 1
		if modNoVar {
			idx = 0
		}
		for i := idx; i < len(chunks); i++ {
			if m := reMod.FindSubmatch(chunks[i]); m != nil {
				fn := GetModFn(byteconv.B2S(m[1]))
				if fn == nil {
					continue
				}
				args := extractArgs(m[2])
				mods = append(mods, mod{
					id:  m[1],
					fn:  fn,
					arg: args,
				})
			}
		}
		return chunks[0], mods
	} else {
		return p, nil
	}
}

// Get list of arguments of modifier or callback, ex:
// variable|mod(arg0, ..., argN)
//
// _____________^          ^
//
// callback(arg0, ..., argN)
//
// _________^          ^
func extractArgs(l []byte) []*arg {
	r := make([]*arg, 0)
	if len(l) == 0 {
		return r
	}
	args := bytes.Split(l, comma)
	for _, a := range args {
		var set [][]byte
		a = bytealg.Trim(a, space)
		static := isStatic(a)
		if !static {
			a, set = extractSet(a)
		} else {
			a = bytealg.Trim(a, quotes)
		}
		r = append(r, &arg{
			val:    a,
			subset: set,
			static: static,
		})
	}
	return r
}

// Get list of certain keys that should be checked sequentially.
func extractSet(p []byte) ([]byte, [][]byte) {
	p = replaceQB(p)
	if m := reSet.FindSubmatch(p); m != nil {
		return m[1], bytes.Split(m[2], vline)
	}
	return p, nil
}

// Replace square brackets in expression like this a[key] -> a.key
func replaceQB(p []byte) []byte {
	p = bytes.Replace(p, qbO, dot, -1)
	p = bytes.Replace(p, qbC, empty, -1)
	return p
}
