package jsondecoder

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/koykov/bytealg"
	"github.com/koykov/fastconv"
)

var (
	nl     = []byte("\n")
	vline  = []byte("|")
	space  = []byte(" ")
	comma  = []byte(",")
	dot    = []byte(".")
	empty  = []byte("")
	qbO    = []byte("[")
	qbC    = []byte("]")
	noFmt  = []byte(" \t\n\r")
	quotes = []byte("\"'`")

	reAssignV2V = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*(.*)`)
	reAssignF2V = regexp.MustCompile(`(?i)([\w\d\\.\[\]]+)\s*=\s*([^(|]+)\(([^)]*)\)`)
	reFunction  = regexp.MustCompile(`([^(]+)\(([^)]*)\)`)
	reMod       = regexp.MustCompile(`([^(]+)\(*([^)]*)\)*`)
	reSet       = regexp.MustCompile(`(.*)\.{([^}]+)}`)
)

func Parse(src []byte) (rules Rules, err error) {
	lines := bytes.Split(src, nl)
	rules = make(Rules, 0, len(lines))
	for i, line := range lines {
		rule := rule{}
		line = bytealg.Trim(line, noFmt)
		if reAssignV2V.Match(line) {
			if m := reAssignF2V.FindSubmatch(line); m != nil {
				rule.dst = replaceQB(m[1])
				rule.src = replaceQB(m[2])
				fn := GetGetterFn(fastconv.B2S(m[2]))
				if fn == nil {
					err = fmt.Errorf("unknown getter function '%s' at line %d", m[2], i)
					break
				}
				rule.getter = fn
				rule.arg = extractArgs(m[3])
			} else if m := reAssignV2V.FindSubmatch(line); m != nil {
				rule.dst = replaceQB(m[1])
				if rule.static = isStatic(m[2]); rule.static {
					rule.src = bytealg.Trim(m[2], quotes)
				} else {
					rule.src, rule.mod = extractMods(m[2])
					rule.src, rule.set = extractSet(rule.src)
				}
			}
			rules = append(rules, rule)
			continue
		}
		if m := reFunction.FindSubmatch(line); m != nil {
			rule.src = m[1]
			fn := GetCallbackFn(fastconv.B2S(m[1]))
			if fn == nil {
				err = fmt.Errorf("unknown callback function '%s' at line %d", m[1], i)
				break
			}
			rule.callback = fn
			rule.arg = extractArgs(m[2])

			rules = append(rules, rule)
			continue
		}
		err = fmt.Errorf("unknown rule '%s' a line %d", line, i)
		break
	}
	return
}

func extractMods(p []byte) ([]byte, []mod) {
	if bytes.Contains(p, vline) {
		// First try to extract suffix mods, like ...|default(0).
		mods := make([]mod, 0)
		chunks := bytes.Split(p, vline)
		for i := 1; i < len(chunks); i++ {
			if m := reMod.FindSubmatch(chunks[i]); m != nil {
				fn := GetModFn(fastconv.B2S(m[1]))
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
			set:    set,
			static: static,
		})
	}
	return r
}

func extractSet(p []byte) ([]byte, [][]byte) {
	p = replaceQB(p)
	if m := reSet.FindSubmatch(p); m != nil {
		return m[1], bytes.Split(m[2], vline)
	}
	return p, nil
}

func replaceQB(p []byte) []byte {
	p = bytes.Replace(p, qbO, dot, -1)
	p = bytes.Replace(p, qbC, empty, -1)
	return p
}
