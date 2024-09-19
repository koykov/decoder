package decoder

const (
	// Types of targets.
	targetCond = iota
	targetLoop
	targetSwitch
)

// Target is a storage of depths needed to provide proper out from conditions, loops and switches control structures.
type target map[int]int

// Create new target based on current parser state.
func newTarget(p *Parser) *target {
	return &target{
		targetCond:   p.cc,
		targetLoop:   p.cl,
		targetSwitch: p.cs,
	}
}

// Check if parser reached the target.
func (t *target) reached(p *Parser) bool {
	return (*t)[targetCond] == p.cc &&
		(*t)[targetLoop] == p.cl &&
		(*t)[targetSwitch] == p.cs
}

// Check if target is a root.
func (t *target) eqZero() bool {
	return (*t)[targetCond] == 0 &&
		(*t)[targetLoop] == 0 &&
		(*t)[targetSwitch] == 0
}
