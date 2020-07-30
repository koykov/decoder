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
	noFmt  = []byte(" \t\n\r")
	quotes = []byte("\"'`")

	reAssignV2V = regexp.MustCompile(`(?i)([\w\d\\.]+)\s*=\s*(.*)`)
	reAssignF2V = regexp.MustCompile(`(?i)([\w\d\\.]+)\s*=\s*([^(|]+)\(([^)]*)\)`)
	reFunction  = regexp.MustCompile(`([^(]+)\(([^)]*)\)`)
	reMod       = regexp.MustCompile(`([^(]+)\(*([^)]*)\)*`)
)

func Parse(src []byte) (rules Rules, err error) {
	lines := bytes.Split(src, nl)
	rules = make(Rules, 0, len(lines))
	for i, line := range lines {
		rule := rule{}
		line = bytealg.Trim(line, noFmt)
		if reAssignV2V.Match(line) {
			if m := reAssignF2V.FindSubmatch(line); m != nil {
				//
			} else if m := reAssignV2V.FindSubmatch(line); m != nil {
				rule.dst = m[1]
				if rule.static = isStatic(m[2]); rule.static {
					rule.src = bytealg.Trim(m[2], quotes)
				} else {
					rule.src, rule.mod = extractMods(m[2])
				}
			}
			rules = append(rules, rule)
			continue
		}
		if m := reFunction.FindSubmatch(line); m != nil {
			rules = append(rules, rule)
			continue
		}
		err = fmt.Errorf("unknown rule '%s' a line %d", line, i)
		break
	}
	return
}

func extractMods(t []byte) ([]byte, []mod) {
	if bytes.Contains(t, vline) {
		// First try to extract suffix mods, like ...|default(0).
		mods := make([]mod, 0)
		chunks := bytes.Split(t, vline)
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
		return t, nil
	}
}

func extractArgs(l []byte) []*arg {
	r := make([]*arg, 0)
	if len(l) == 0 {
		return r
	}
	args := bytes.Split(l, comma)
	for _, a := range args {
		a = bytealg.Trim(a, space)
		r = append(r, &arg{
			val:    bytealg.Trim(a, quotes),
			static: isStatic(a),
		})
	}
	return r
}
