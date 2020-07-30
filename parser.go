package jsondecoder

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/koykov/bytealg"
)

var (
	nl     = []byte("\n")
	noFmt  = []byte(" \t\n\r")
	quotes = []byte("\"'`")

	reAssignV2V = regexp.MustCompile(`(?i)([\w\d\\.]+)\s*=\s*(.*)`)
	reAssignF2V = regexp.MustCompile(`(?i)([\w\d\\.]+)\s*=\s*([^(]+)\(([^)]*)\)`)
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
				rule.src = bytealg.Trim(m[2], quotes)
				rule.static = isStatic(m[2])
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
