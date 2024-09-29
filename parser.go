package decoder

import (
	"bytes"
	"fmt"
	"hash/crc64"
	"os"
	"regexp"
	"strconv"

	"github.com/koykov/bytealg"
	"github.com/koykov/byteconv"
)

type parser struct {
	target

	// Decoder body to parse.
	body []byte
}

var (
	// Byte constants.
	nl       = []byte("\n")
	vline    = []byte("|")
	space    = []byte(" ")
	comma    = []byte(",")
	uscore   = []byte("_")
	dot      = []byte(".")
	empty    = []byte("")
	qbO      = []byte("[")
	qbC      = []byte("]")
	noFmt    = []byte(" \t\n\r")
	quotes   = []byte("\"'`")
	comment  = []byte("//")
	loopBrk  = []byte("break")
	loopLBrk = []byte("lazybreak")
	loopCont = []byte("continue")
	_, _     = nl, noFmt

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
	reLoopBrk   = regexp.MustCompile(`break (\d+)`)
	reLoopLBrk  = regexp.MustCompile(`lazybreak (\d+)`)

	crc64Tab = crc64.MakeTable(crc64.ISO)

	// Suppress go vet warning.
	_ = ParseFile
)

// Parse parses the decoder rules.
func Parse(src []byte) (*Tree, error) {
	p := &parser{body: src}
	hsum := crc64.Checksum(p.body, crc64Tab)
	if tree := decDB.getTreeByHash(hsum); tree != nil {
		return tree, nil
	}

	t := p.targetSnapshot()
	nodes, _, err := p.parse(nil, nil, 0, t)
	return &Tree{
		nodes: nodes,
		hsum:  0,
	}, err
}

// ParseFile parses the file.
func ParseFile(fileName string) (tree *Tree, err error) {
	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	var raw []byte
	raw, err = os.ReadFile(fileName)
	if err != nil {
		return
	}
	return Parse(raw)
}

func (p *parser) parse(dst []node, root *node, offset int, t *target) ([]node, int, error) {
	var (
		ctl     []byte
		eof, up bool
		err     error
	)
	for !t.reached(p) || t.eqZero() {
		if ctl, offset, eof = p.nextCtl(offset); eof {
			break
		}
		if len(ctl) == 0 {
			continue
		}
		r := node{typ: typeOperator}
		if dst, offset, up, err = p.processCtl(dst, root, &r, ctl, offset); err != nil {
			return dst, offset, err
		}
		if up {
			break
		}
	}
	if !t.reached(p) {
		err = ErrUnbalancedCtl
	}
	return dst, offset, err
}

func (p *parser) nextCtl(offset int) ([]byte, int, bool) {
	var eof bool
	if offset, eof = p.skipFmt(offset); eof {
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

func (p *parser) processCtl(dst []node, root, r *node, ctl []byte, offset int) ([]node, int, bool, error) {
	var err error
	switch {
	case ctl[0] == '#' || bytes.HasPrefix(ctl, comment):
		offset += len(ctl)
		return dst, offset, false, nil
	case reLoop.Match(ctl):
		if m := reLoopRange.FindSubmatch(ctl); m != nil {
			r.typ = typeLoopRange
			if bytes.Contains(m[1], comma) {
				kv := bytes.Split(m[1], comma)
				r.loopKey = bytealg.Trim(kv[0], space)
				if bytes.Equal(r.loopKey, uscore) {
					r.loopKey = nil
				}
				r.loopVal = bytealg.Trim(kv[1], space)
			} else {
				r.loopKey = bytealg.Trim(m[1], space)
			}
			r.loopSrc = m[2]
		} else if m := reLoopCount.FindSubmatch(ctl); m != nil {
			r.typ = typeLoopCount
			r.loopCnt = m[1]
			r.loopCntInit = m[2]
			r.loopCntStatic = isStatic(m[2])
			r.loopCondOp = p.parseOp(m[3])
			r.loopLim = m[4]
			r.loopLimStatic = isStatic(m[4])
			r.loopCntOp = p.parseOp(m[5])
		} else {
			return dst, offset, false, fmt.Errorf("couldn't parse loop control structure '%s' at offset %d", string(ctl), offset)
		}
		t := p.targetSnapshot()
		p.cl++

		offset += len(ctl)
		if r.child, offset, err = p.parse(r.child, r, offset, t); err != nil {
			return dst, offset, false, err
		}
		dst = append(dst, *r)
	case ctl[0] == '}':
		offset++
		switch root.typ {
		case typeLoopCount, typeLoopRange:
			p.cl--
		default:
			// todo check other cases
		}
		return dst, offset, true, err
	case reLoopLBrk.Match(ctl):
		r.typ = typeLBreak
		m := reLoopLBrk.FindSubmatch(ctl)
		if i, _ := strconv.ParseInt(byteconv.B2S(m[1]), 10, 64); i > 0 {
			r.loopBrkD = int(i)
		}
		dst = append(dst, *r)
		offset += len(ctl)
	case bytes.Equal(ctl, loopLBrk):
		r.typ = typeLBreak
		dst = append(dst, *r)
		offset += len(ctl)
	case reLoopBrk.Match(ctl):
		r.typ = typeBreak
		m := reLoopBrk.FindSubmatch(ctl)
		if i, _ := strconv.ParseInt(byteconv.B2S(m[1]), 10, 64); i > 0 {
			r.loopBrkD = int(i)
		}
		dst = append(dst, *r)
		offset += len(ctl)
	case bytes.Equal(ctl, loopBrk):
		r.typ = typeBreak
		dst = append(dst, *r)
		offset += len(ctl)
	case bytes.Equal(ctl, loopCont):
		r.typ = typeContinue
		dst = append(dst, *r)
		offset += len(ctl)
	case reAssignV2C.Match(ctl):
		// Var-to-ctx expression caught.
		if m := reAssignV2CAs.FindSubmatch(ctl); m != nil {
			r.dst = m[1]
			r.src = m[2]
			r.ins = m[3]
		} else if m = reAssignV2CDot.FindSubmatch(ctl); m != nil {
			r.dst = m[1]
			r.src = m[2]
			r.ins = m[3]
		} else if m = reAssignV2C.FindSubmatch(ctl); m != nil {
			r.dst = m[1]
			r.src = m[2]
		}
		// Check static/variable.
		if r.static = isStatic(r.src); r.static {
			r.src = bytealg.Trim(r.src, quotes)
		} else {
			r.src, r.mod = extractMods(r.src)
			r.src, r.subset = extractSet(r.src)
		}
		dst = append(dst, *r)
		offset += len(ctl)
	case reAssignV2V.Match(ctl):
		// Var-to-var expression caught.
		if m := reAssignF2V.FindSubmatch(ctl); m != nil {
			// Func-to-var expression caught.
			r.dst = replaceQB(m[1])
			r.src = replaceQB(m[2])
			// Parse getter callback.
			fn := GetGetterFn(byteconv.B2S(m[2]))
			if fn != nil {
				r.getter = fn
				r.arg = extractArgs(m[3])
			} else {
				// Getter func not found, so try to fallback to mod func.
				m = reAssignV2V.FindSubmatch(ctl)
				r.src, r.mod = extractMods(m[2])
				r.src, r.subset = extractSet(r.src)
			}
			if r.getter == nil && len(r.mod) == 0 {
				err = fmt.Errorf("unknown getter nor modifier function '%s' at offset %d", m[2], offset)
				break
			}
		} else if m = reAssignV2V.FindSubmatch(ctl); m != nil {
			// Var-to-var ...
			r.dst = replaceQB(m[1])
			if r.static = isStatic(m[2]); r.static {
				r.src = bytealg.Trim(m[2], quotes)
			} else {
				r.src, r.mod = extractMods(m[2])
				r.src, r.subset = extractSet(r.src)
			}
		}
		dst = append(dst, *r)
		offset += len(ctl)
	case reFunction.Match(ctl):
		m := reFunction.FindSubmatch(ctl)
		// Function expression caught.
		r.src = m[1]
		// Parse callback.
		fn := GetCallbackFn(byteconv.B2S(m[1]))
		if fn == nil {
			err = fmt.Errorf("unknown callback function '%s' at offset %d", m[1], offset)
			break
		}
		r.callback = fn
		r.arg = extractArgs(m[2])

		dst = append(dst, *r)
		offset += len(ctl)
	default:
		return dst, offset, false, fmt.Errorf("unknown node '%s' at position %d", string(ctl), offset)
	}
	return dst, offset, false, err
}

func (p *parser) skipFmt(offset int) (int, bool) {
	n := len(p.body)
	for i := offset; i < n; i++ {
		c := p.body[i]
		if c != '\n' && c != '\r' && c != '\t' && c != ' ' && c != ';' {
			return i, i == n-1
		}
	}
	return n - 1, true
}

func (p *parser) parseOp(src []byte) Op {
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

func (p *parser) targetSnapshot() *target {
	cpy := p.target
	return &cpy
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
