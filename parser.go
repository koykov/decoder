package decoder

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/koykov/bytealg"
	"github.com/koykov/fastconv"
)

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

	// Suppress go vet warning.
	_ = ParseFile
)

// Parse parses the decoder rules.
func Parse(src []byte) (ruleset Ruleset, err error) {
	// Split body to separate lines.
	// Each line contains only one expression.
	lines := bytes.Split(src, nl)
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
		rule := rule{}
		if reAssignV2C.Match(line) {
			// Var-to-ctx expression caught.
			if m := reAssignV2CAs.FindSubmatch(line); m != nil {
				rule.dst = m[1]
				rule.src = m[2]
				rule.ins = m[3]
			} else if m := reAssignV2CDot.FindSubmatch(line); m != nil {
				rule.dst = m[1]
				rule.src = m[2]
				rule.ins = m[3]
			} else if m := reAssignV2C.FindSubmatch(line); m != nil {
				rule.dst = m[1]
				rule.src = m[2]
			}
			// Check static/variable.
			if rule.static = isStatic(rule.src); rule.static {
				rule.src = bytealg.Trim(rule.src, quotes)
			} else {
				rule.src, rule.mod = extractMods(rule.src)
				rule.src, rule.subset = extractSet(rule.src)
			}
			ruleset = append(ruleset, rule)
			continue
		}
		if reAssignV2V.Match(line) {
			// Var-to-var expression caught.
			if m := reAssignF2V.FindSubmatch(line); m != nil {
				// Func-to-var expression caught.
				rule.dst = replaceQB(m[1])
				rule.src = replaceQB(m[2])
				// Parse getter callback.
				fn := GetGetterFn(fastconv.B2S(m[2]))
				if fn == nil {
					err = fmt.Errorf("unknown getter function '%s' at line %d", m[2], i)
					break
				}
				rule.getter = fn
				rule.arg = extractArgs(m[3])
			} else if m := reAssignV2V.FindSubmatch(line); m != nil {
				// Var-to-var ...
				rule.dst = replaceQB(m[1])
				if rule.static = isStatic(m[2]); rule.static {
					rule.src = bytealg.Trim(m[2], quotes)
				} else {
					rule.src, rule.mod = extractMods(m[2])
					rule.src, rule.subset = extractSet(rule.src)
				}
			}
			ruleset = append(ruleset, rule)
			continue
		}
		if m := reFunction.FindSubmatch(line); m != nil {
			// Function expression caught.
			rule.src = m[1]
			// Parse callback.
			fn := GetCallbackFn(fastconv.B2S(m[1]))
			if fn == nil {
				err = fmt.Errorf("unknown callback function '%s' at line %d", m[1], i)
				break
			}
			rule.callback = fn
			rule.arg = extractArgs(m[2])

			ruleset = append(ruleset, rule)
			continue
		}
		// Report unparsed error.
		err = fmt.Errorf("unknown rule '%s' a line %d", line, i)
		break
	}
	return
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
