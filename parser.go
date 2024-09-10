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
	dec []byte

	// Counters (depths) of conditions, loops and switches.
	cc, cl, cs int
}

var (
	// Byte constants.
	nl      = []byte("\n")
	vline   = []byte("|")
	space   = []byte(" ")
	comma   = []byte(",")
	dot     = []byte(".")
	empty   = []byte("")
	qbO     = []byte("[")
	qbC     = []byte("]")
	noFmt   = []byte(" \t\n\r")
	quotes  = []byte("\"'`")
	comment = []byte("//")

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
func Parse(src []byte) (ruleset Ruleset, err error) {
	p := &Parser{dec: src}
	return p.parse()
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

func (p *Parser) parse() (ruleset Ruleset, err error) {
	// Split body to separate lines.
	// Each line contains only one expression.
	lines := bytes.Split(p.dec, nl)
	ruleset = make(Ruleset, 0, len(lines))
	for i, line := range lines {
		if len(line) == 0 || (len(line) > 1 && bytes.Equal(line[:2], comment)) {
			continue
		}
		line = bytealg.Trim(line, noFmt)
		if len(line) == 0 || line[0] == '#' {
			// Ignore comments.
			continue
		}
		r := rule{typ: typeOperator}
		switch {
		case reAssignV2C.Match(line):
			// Var-to-ctx expression caught.
			if m := reAssignV2CAs.FindSubmatch(line); m != nil {
				r.dst = m[1]
				r.src = m[2]
				r.ins = m[3]
			} else if m = reAssignV2CDot.FindSubmatch(line); m != nil {
				r.dst = m[1]
				r.src = m[2]
				r.ins = m[3]
			} else if m = reAssignV2C.FindSubmatch(line); m != nil {
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
			ruleset = append(ruleset, r)
			continue
		case reAssignV2V.Match(line):
			// Var-to-var expression caught.
			if m := reAssignF2V.FindSubmatch(line); m != nil {
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
					m = reAssignV2V.FindSubmatch(line)
					r.src, r.mod = extractMods(m[2])
					r.src, r.subset = extractSet(r.src)
				}
				if r.getter == nil && len(r.mod) == 0 {
					err = fmt.Errorf("unknown getter nor modifier function '%s' at line %d", m[2], i)
					break
				}
			} else if m = reAssignV2V.FindSubmatch(line); m != nil {
				// Var-to-var ...
				r.dst = replaceQB(m[1])
				if r.static = isStatic(m[2]); r.static {
					r.src = bytealg.Trim(m[2], quotes)
				} else {
					r.src, r.mod = extractMods(m[2])
					r.src, r.subset = extractSet(r.src)
				}
			}
			ruleset = append(ruleset, r)
			continue
		case reFunction.Match(line):
			m := reFunction.FindSubmatch(line)
			// Function expression caught.
			r.src = m[1]
			// Parse callback.
			fn := GetCallbackFn(byteconv.B2S(m[1]))
			if fn == nil {
				err = fmt.Errorf("unknown callback function '%s' at line %d", m[1], i)
				break
			}
			r.callback = fn
			r.arg = extractArgs(m[2])

			ruleset = append(ruleset, r)
			continue
		case reLoop.Match(line):
			if m := reLoopRange.FindSubmatch(line); m != nil {
				r.typ = typeLoopRange
				// todo implement me
			} else if m := reLoopCount.FindSubmatch(line); m != nil {
				r.typ = typeLoopCount
				// todo implement me
			} else {
				return ruleset, fmt.Errorf("couldn't parse loop control structure '%s' at offset %d", line, i)
			}
			t := newTarget(p)
			_ = t
			p.cl++

			if r.child, err = p.parse(); err != nil {
				return ruleset, err
			}
			ruleset = append(ruleset, r)
			continue
		case true:
			// todo check loop end
		}
		// Report unparsed error.
		err = fmt.Errorf("unknown rule '%s' a line %d", line, i)
		break
	}
	return
}

func (p *Parser) parse1() (ruleset Ruleset, err error) {
	// todo implement me
	return
}

func (p *Parser) nextCtl(offset int) ([]byte, int, bool) {
	// todo implement me
	return nil, offset, false
}

func (p *Parser) processCtl(ruleset Ruleset, root *rule, ctl []byte, pos int) ([]Ruleset, int, bool, error) {
	_, _, _, _ = ruleset, root, ctl, pos
	// todo implement me
	return nil, 0, false, nil
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
