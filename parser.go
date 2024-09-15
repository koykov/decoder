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
	ruleset, _, err := p.parse1(nil, nil, 0, t)
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

func (p *Parser) parse1(dst Ruleset, root *rule, offset int, t *target) (Ruleset, int, error) {
	var (
		ctl     []byte
		eol, up bool
		err     error
	)
	for !t.reached(p) || t.eqZero() {
		ctl, offset, eol = p.nextCtl(offset)
		_ = eol
		if len(ctl) == 0 {
			continue
		}
		r := rule{typ: typeOperator}
		if dst, offset, up, err = p.processCtl(dst, root, &r, ctl, offset); err != nil {
			return dst, offset, err
		}
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
	if j := bytes.IndexByte(ctl, ';'); j > 0 && !bytes.HasPrefix(ctl, []byte("for")) {
		return p.body[offset : offset+j], offset, false
	}
	if j := bytes.IndexByte(ctl, '}'); j > 0 && bytes.Index(ctl, []byte(".{")) == -1 {
		return p.body[offset : offset+j], offset, false
	}
	return p.body[offset : offset+i], offset, false
}

func (p *Parser) processCtl(dst Ruleset, root, node *rule, ctl []byte, offset int) (Ruleset, int, bool, error) {
	var err error
	switch {
	case ctl[0] == '#' || bytes.HasPrefix(ctl, comment):
		offset += len(ctl)
		return dst, offset, false, nil
	case reLoop.Match(ctl):
		if m := reLoopRange.FindSubmatch(ctl); m != nil {
			node.typ = typeLoopRange
			if bytes.Contains(m[1], comma) {
				kv := bytes.Split(m[1], comma)
				node.loopKey = bytealg.Trim(kv[0], space)
				if bytes.Equal(node.loopKey, uscore) {
					node.loopKey = nil
				}
				node.loopVal = bytealg.Trim(kv[1], space)
			} else {
				node.loopKey = bytealg.Trim(m[1], space)
			}
			node.loopSrc = m[2]
		} else if m := reLoopCount.FindSubmatch(ctl); m != nil {
			node.typ = typeLoopCount
			node.loopCnt = m[1]
			node.loopCntInit = m[2]
			node.loopCntStatic = isStatic(m[2])
			node.loopCondOp = p.parseOp(m[3])
			node.loopLim = m[4]
			node.loopLimStatic = isStatic(m[4])
			node.loopCntOp = p.parseOp(m[5])
		} else {
			return dst, offset, false, fmt.Errorf("couldn't parse loop control structure '%s' at offset %d", string(ctl), offset)
		}
		t := newTarget(p)
		p.cl++

		offset += len(ctl)
		if node.child, offset, err = p.parse1(node.child, node, offset, t); err != nil {
			return dst, offset, false, err
		}
		dst = append(dst, *node)
	case ctl[0] == '}':
		offset++
		switch root.typ {
		case typeLoopCount, typeLoopRange:
			p.cl--
		default:
			// todo check other cases
		}
		return dst, offset, true, err
	case reAssignV2C.Match(ctl):
		// Var-to-ctx expression caught.
		if m := reAssignV2CAs.FindSubmatch(ctl); m != nil {
			node.dst = m[1]
			node.src = m[2]
			node.ins = m[3]
		} else if m = reAssignV2CDot.FindSubmatch(ctl); m != nil {
			node.dst = m[1]
			node.src = m[2]
			node.ins = m[3]
		} else if m = reAssignV2C.FindSubmatch(ctl); m != nil {
			node.dst = m[1]
			node.src = m[2]
		}
		// Check static/variable.
		if node.static = isStatic(node.src); node.static {
			node.src = bytealg.Trim(node.src, quotes)
		} else {
			node.src, node.mod = extractMods(node.src)
			node.src, node.subset = extractSet(node.src)
		}
		dst = append(dst, *node)
		offset += len(ctl)
	case reAssignV2V.Match(ctl):
		// Var-to-var expression caught.
		if m := reAssignF2V.FindSubmatch(ctl); m != nil {
			// Func-to-var expression caught.
			node.dst = replaceQB(m[1])
			node.src = replaceQB(m[2])
			// Parse getter callback.
			fn := GetGetterFn(byteconv.B2S(m[2]))
			if fn != nil {
				node.getter = fn
				node.arg = extractArgs(m[3])
			} else {
				// Getter func not found, so try to fallback to mod func.
				m = reAssignV2V.FindSubmatch(ctl)
				node.src, node.mod = extractMods(m[2])
				node.src, node.subset = extractSet(node.src)
			}
			if node.getter == nil && len(node.mod) == 0 {
				err = fmt.Errorf("unknown getter nor modifier function '%s' at offset %d", m[2], offset)
				break
			}
		} else if m = reAssignV2V.FindSubmatch(ctl); m != nil {
			// Var-to-var ...
			node.dst = replaceQB(m[1])
			if node.static = isStatic(m[2]); node.static {
				node.src = bytealg.Trim(m[2], quotes)
			} else {
				node.src, node.mod = extractMods(m[2])
				node.src, node.subset = extractSet(node.src)
			}
		}
		dst = append(dst, *node)
		offset += len(ctl)
	case reFunction.Match(ctl):
		m := reFunction.FindSubmatch(ctl)
		// Function expression caught.
		node.src = m[1]
		// Parse callback.
		fn := GetCallbackFn(byteconv.B2S(m[1]))
		if fn == nil {
			err = fmt.Errorf("unknown callback function '%s' at offset %d", m[1], offset)
			break
		}
		node.callback = fn
		node.arg = extractArgs(m[2])

		dst = append(dst, *node)
		offset += len(ctl)
	default:
		return dst, offset, false, fmt.Errorf("unknown rule '%s' at position %d", string(ctl), offset)
	}
	return dst, offset, false, err
}

func (p *Parser) skipFmt(offset int) (int, bool) {
	n := len(p.body)
	for i := offset; i < n; i++ {
		c := p.body[i]
		if c != '\n' && c != '\r' && c != '\t' && c != ' ' && c != ';' {
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
