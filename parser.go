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
	condLen  = []byte("len")
	condCap  = []byte("cap")
	_, _     = nl, noFmt

	// Operation constants.
	opEq_  = []byte("==")
	opNq_  = []byte("!=")
	opGt_  = []byte(">")
	opGtq_ = []byte(">=")
	opLt_  = []byte("<")
	opLtq_ = []byte("<=")
	opInc_ = []byte("++")
	opDec_ = []byte("--")

	// Regexp to parse expressions.
	reAssignV2CAs  = regexp.MustCompile(`((?:context|ctx)\.[\w\d\\.\[\]]+)\s*=\s*(.*) as ([:\w]*)`)
	reAssignV2CDot = regexp.MustCompile(`((?:context|ctx)\.[\w\d\\.\[\]]+)\s*=\s*(.*).\(([:\w]*)\)`)
	reAssignV2C    = regexp.MustCompile(`((?:context|ctx)\.[\w\d\\.\[\]]+)\s*=\s*(.*)`)

	reAssignV2V = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*(.*)`)
	reAssignF2V = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*([^(|]+)\(([^)]*)\)`)
	reFunction  = regexp.MustCompile(`([^(]+)\(([^)]*)\)`)
	reMod       = regexp.MustCompile(`([^(]+)\(*([^)]*)\)*`)
	reSet       = regexp.MustCompile(`(.*)\.{([^}]+)}`)

	reTernary         = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*(.*)(==|!=|>=|<=|>|<)(.*)\s*\?\s*([^:]+):(.*)`)
	reTernaryHelper   = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*([^(]+)\(*([^)]*)\)\s*\?\s*([^:]+):(.*)`)
	reTernaryCondExpr = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*(.*)\s*(==|!=|>=|<=|>|<)([^?]+)`)

	reLoop      = regexp.MustCompile(`for .*`)
	reLoopRange = regexp.MustCompile(`for ([^:]+)\s*:*=\s*range\s*([^\s]*)\s*\{` + "")
	reLoopCount = regexp.MustCompile(`for (\w*)\s*:*=\s*(\w+)\s*;\s*\w+\s*(<|<=|>|>=|!=)+\s*([^;]+)\s*;\s*\w*(--|\+\+)+\s*\{`)
	reLoopBrk   = regexp.MustCompile(`break (\d+)`)
	reLoopLBrk  = regexp.MustCompile(`lazybreak (\d+)`)

	reCond        = regexp.MustCompile(`if .*`)
	reCondExpr    = regexp.MustCompile(`if (.*)(==|!=|>=|<=|>|<)(.*)\s*{`)
	reCondHelper  = regexp.MustCompile(`if ([^(]+)\(*([^)]*)\)\s*{`)
	reCondComplex = regexp.MustCompile(`if .*&&|\|\||\(|\).*\s*{`)
	reCondOK      = regexp.MustCompile(`if (\w+),*\s*(\w*)\s*:*=\s*([^(]+)\(*([^)]*)\)(.*)\s*;\s*([!\w]+)\s*{`)
	reCondAsOK    = regexp.MustCompile(`if (\w+),*\s*(\w*)\s*:*=\s*([^(]+)\(*([^)]*)\) as (\w*)\s*;\s*([!\w]+)\s*{`)
	reCondDotOK   = regexp.MustCompile(`if (\w+),*\s*(\w*)\s*:*=\s*([^(]+)\(*([^)]*)\)\.\((\w*)\)\s*;\s*([!\w]+)\s*{`)
	reCondExprOK  = regexp.MustCompile(`if .*;\s*([!:\w]+)(.*)(.*)\s*{`)
	reCondElse    = regexp.MustCompile(`}\s*else\s*{`)

	reSwitch           = regexp.MustCompile(`^switch\s*([^\s^{]*)\s*{`)
	reSwitchCase       = regexp.MustCompile(`case ([^<=>!]+)([<=>!]{2})*(.*):`)
	reSwitchCaseHelper = regexp.MustCompile(`case ([^(]+)\(*([^)]*)\):`)
	reSwitchDefault    = regexp.MustCompile(`default\s*:`)

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
	if ctl[0] == '#' || bytes.HasPrefix(ctl, comment) {
		offset += len(ctl)
		return dst, offset, false, nil
	}
	if reLoop.Match(ctl) {
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
		return dst, offset, false, err
	}

	if reCondOK.Match(ctl) {
		r.typ = typeCondOK
		var m [][]byte
		m = reCondAsOK.FindSubmatch(ctl)
		if m == nil {
			m = reCondDotOK.FindSubmatch(ctl)
		}
		if m == nil {
			m = reCondOK.FindSubmatch(ctl)
		}
		r.condOKL, r.condOKR = m[1], m[2]
		r.condHlp, r.condHlpArg = m[3], extractArgs(m[4])
		if len(m[5]) > 0 {
			r.condIns = m[5]
		}
		r.condL, r.condR, r.condStaticL, r.condStaticR, r.condOp = p.parseCondExpr(reCondExprOK, ctl)

		t := p.targetSnapshot()
		p.cc++

		var subNodes []node
		offset += len(ctl)
		subNodes, offset, err = p.parse(subNodes, &node{typ: typeCondOK}, offset, t)
		split := splitNodes(subNodes)
		if len(split) > 0 {
			nodeTrue := node{typ: typeCondTrue, child: split[0]}
			r.child = append(r.child, nodeTrue)
		}
		if len(split) > 1 {
			nodeFalse := node{typ: typeCondFalse, child: split[1]}
			r.child = append(r.child, nodeFalse)
		}

		dst = append(dst, *r)
		return dst, offset, false, err
	}
	if reCond.Match(ctl) {
		dst, offset, err = p.processCond(dst, r, ctl, offset)
		return dst, offset, false, err
	}
	if reCondElse.Match(ctl) {
		root.typ = typeDiv
		dst = append(dst, *root)
		offset += len(ctl)
		return dst, offset, false, err
	}

	if m := reSwitch.FindSubmatch(ctl); m != nil {
		// Create new target, increase switch counter and dive deeper.
		t := p.targetSnapshot()
		p.cs++

		r.typ = typeSwitch
		if len(m) > 0 {
			r.switchArg = m[1]
		}
		r.child = make([]node, 0)
		r.child, offset, err = p.parse(r.child, r, offset+len(ctl), t)
		r.child = rollupSwitchNodes(r.child)

		dst = append(dst, *r)
		return dst, offset, false, err
	}
	// Check switch's case with condition helper.
	if m := reSwitchCaseHelper.FindSubmatch(ctl); m != nil {
		r.typ = typeCase
		r.caseHlp = m[1]
		r.caseHlpArg = extractArgs(m[2])
		dst = append(dst, *r)
		offset = offset + len(ctl)
		return dst, offset, false, err
	}
	// Check switch's case with simple condition.
	if reSwitchCase.Match(ctl) {
		r.typ = typeCase
		r.caseL, r.caseR, r.caseStaticL, r.caseStaticR, r.caseOp = p.parseCaseExpr(ctl)
		dst = append(dst, *r)
		offset = offset + len(ctl)
		return dst, offset, false, err
	}
	// Check switch's default.
	if reSwitchDefault.Match(ctl) {
		r.typ = typeDefault
		dst = append(dst, *r)
		offset = offset + len(ctl)
		return dst, offset, false, err
	}

	if ctl[0] == '}' {
		offset++
		switch root.typ {
		case typeLoopCount, typeLoopRange:
			p.cl--
		case typeCond, typeCondOK, typeElse, typeDiv:
			p.cc--
		case typeSwitch:
			p.cs--
		default:
			err = ErrUnexpectedClose
		}
		return dst, offset, true, err
	}
	if reLoopLBrk.Match(ctl) {
		r.typ = typeLBreak
		m := reLoopLBrk.FindSubmatch(ctl)
		if i, _ := strconv.ParseInt(byteconv.B2S(m[1]), 10, 64); i > 0 {
			r.loopBrkD = int(i)
		}
		dst = append(dst, *r)
		offset += len(ctl)
		return dst, offset, false, err
	}
	if bytes.Equal(ctl, loopLBrk) {
		r.typ = typeLBreak
		dst = append(dst, *r)
		offset += len(ctl)
		return dst, offset, false, err
	}
	if reLoopBrk.Match(ctl) {
		r.typ = typeBreak
		m := reLoopBrk.FindSubmatch(ctl)
		if i, _ := strconv.ParseInt(byteconv.B2S(m[1]), 10, 64); i > 0 {
			r.loopBrkD = int(i)
		}
		dst = append(dst, *r)
		offset += len(ctl)
		return dst, offset, false, err
	}
	if bytes.Equal(ctl, loopBrk) {
		r.typ = typeBreak
		dst = append(dst, *r)
		offset += len(ctl)
		return dst, offset, false, err
	}
	if bytes.Equal(ctl, loopCont) {
		r.typ = typeContinue
		dst = append(dst, *r)
		offset += len(ctl)
		return dst, offset, false, err
	}
	if reAssignV2C.Match(ctl) {
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
		return dst, offset, false, err
	}
	if reAssignV2V.Match(ctl) {
		// Var-to-var expression caught.
		var m [][]byte
		if m = reTernary.FindSubmatch(ctl); m != nil {
			// todo implement me
		} else if m = reTernaryHelper.FindSubmatch(ctl); m != nil {
			// todo implement me
		} else if m = reAssignF2V.FindSubmatch(ctl); m != nil {
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
				return dst, offset, false, err
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
		return dst, offset, false, err
	}
	if reFunction.Match(ctl) {
		m := reFunction.FindSubmatch(ctl)
		// Function expression caught.
		r.src = m[1]
		// Parse callback.
		fn := GetCallbackFn(byteconv.B2S(m[1]))
		if fn == nil {
			err = fmt.Errorf("unknown callback function '%s' at offset %d", m[1], offset)
			return dst, offset, false, err
		}
		r.callback = fn
		r.arg = extractArgs(m[2])

		dst = append(dst, *r)
		offset += len(ctl)
		return dst, offset, false, err
	}
	return dst, offset, false, fmt.Errorf("unknown node '%s' at position %d", string(ctl), offset)
}

func (p *parser) processCond(nodes []node, root *node, ctl []byte, offset int) ([]node, int, error) {
	var (
		subNodes []node
		split    [][]node
		err      error
		pos      = offset
	)
	// Check complexity of the condition first.
	if reCondComplex.Match(ctl) {
		// Check if condition may be handled by the condition helper.
		if m := reCondHelper.FindSubmatch(ctl); m != nil {
			root.typ = typeCond
			root.condHlp = m[1]
			root.condHlpArg = extractArgs(m[2])
			root.condL, root.condR, root.condStaticL, root.condStaticR, root.condOp = p.parseCondExpr(reCondExpr, ctl)
			switch {
			case bytes.Equal(root.condHlp, condLen):
				root.condLC = lcLen
			case bytes.Equal(root.condHlp, condCap):
				root.condLC = lcCap
			}

			t := p.targetSnapshot()
			p.cc++
			subNodes, offset, err = p.parse(subNodes, root, pos+len(ctl), t)
			split = splitNodes(subNodes)

			if len(split) > 0 {
				nodeTrue := node{typ: typeCondTrue, child: split[0]}
				root.child = append(root.child, nodeTrue)
			}
			if len(split) > 1 {
				nodeFalse := node{typ: typeCondFalse, child: split[1]}
				root.child = append(root.child, nodeFalse)
			}
			nodes = append(nodes, *root)
			return nodes, offset, err
		}
		return nodes, pos, fmt.Errorf("too complex condition '%s' at offset %d", ctl, pos)
	}
	root.typ = typeCond
	root.condL, root.condR, root.condStaticL, root.condStaticR, root.condOp = p.parseCondExpr(reCondExpr, ctl)

	// Create new target, increase condition counter and dive deeper.
	t := p.targetSnapshot()
	p.cc++

	subNodes, offset, err = p.parse(subNodes, &node{typ: typeCond}, pos+len(ctl), t)
	split = splitNodes(subNodes)

	if len(split) > 0 {
		nodeTrue := node{typ: typeCondTrue, child: split[0]}
		root.child = append(root.child, nodeTrue)
	}
	if len(split) > 1 {
		nodeFalse := node{typ: typeCondFalse, child: split[1]}
		root.child = append(root.child, nodeFalse)
	}
	nodes = append(nodes, *root)
	return nodes, offset, err
}

// Parse condition to left/right parts and condition operator.
func (p *parser) parseCondExpr(re *regexp.Regexp, expr []byte) (l, r []byte, sl, sr bool, op op) {
	if m := re.FindSubmatch(expr); m != nil {
		l = bytealg.Trim(m[1], space)
		if len(l) > 0 && l[0] == '!' {
			l = l[1:]
			r = bTrue
			sl = false
			sr = true
			op = opNq
		} else {
			r = bytealg.Trim(m[3], space)
			sl = isStatic(l)
			sr = isStatic(r)
			op = p.parseOp(m[2])
		}
		if len(l) > 0 {
			l = bytealg.Trim(l, quotes)
		}
		if len(r) > 0 {
			r = bytealg.Trim(r, quotes)
		}
	}
	return
}

// Parse case condition similar to condition parsing.
func (p *parser) parseCaseExpr(expr []byte) (l, r []byte, sl, sr bool, op op) {
	if m := reSwitchCase.FindSubmatch(expr); m != nil {
		l = bytealg.Trim(m[1], space)
		sl = isStatic(l)
		if len(m) > 1 {
			op = p.parseOp(m[2])
			r = bytealg.Trim(m[3], space)
			sr = isStatic(r)
		}
	}
	return
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

func (p *parser) parseOp(src []byte) op {
	var op_ op
	switch {
	case bytes.Equal(src, opEq_):
		op_ = opEq
	case bytes.Equal(src, opNq_):
		op_ = opNq
	case bytes.Equal(src, opGt_):
		op_ = opGt
	case bytes.Equal(src, opGtq_):
		op_ = opGtq
	case bytes.Equal(src, opLt_):
		op_ = opLt
	case bytes.Equal(src, opLtq_):
		op_ = opLtq
	case bytes.Equal(src, opInc_):
		op_ = opInc
	case bytes.Equal(src, opDec_):
		op_ = opDec
	default:
		op_ = opUnk
	}
	return op_
}

// Split nodes by divider node.
func splitNodes(nodes []node) [][]node {
	if len(nodes) == 0 {
		return nil
	}
	split := make([][]node, 0)
	var o int
	for i, n := range nodes {
		if n.typ == typeDiv {
			split = append(split, nodes[o:i])
			o = i + 1
		}
	}
	if o < len(nodes) {
		split = append(split, nodes[o:])
	}
	return split
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

func rollupSwitchNodes(nodes []node) []node {
	if len(nodes) == 0 {
		return nil
	}
	var (
		r     = make([]node, 0)
		group = node{typ: -1}
	)
	for _, n := range nodes {
		if n.typ != typeCase && n.typ != typeDefault && group.typ == -1 {
			continue
		}
		if n.typ == typeCase || n.typ == typeDefault {
			if group.typ != -1 {
				r = append(r, group)
			}
			group = n
			continue
		}
		group.child = append(group.child, n)
	}
	if len(group.child) > 0 {
		r = append(r, group)
	}
	return r
}
