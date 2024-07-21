package decoder

// ModFn represents signature of the modifier functions.
//
// Arguments description:
// * ctx provides access to additional variables and various buffers to reduce allocations.
// * buf is a storage for final result after finishing modifier work.
// * val is a left side variable that preceded to call of modifier func, example: {%= val|mod(...) %}
// * args is a list of all arguments listed on modifier call.
type ModFn func(ctx *Ctx, buf *any, val any, args []any) error

type ModFnTuple struct {
	docgen
	fn ModFn
}

// Internal modifier representation.
type mod struct {
	id  []byte
	fn  ModFn
	arg []*arg
}

var (
	// Registry of modifiers.
	modRegistry = map[string]int{}
	modBuf      []ModFnTuple

	_ = RegisterModFnNS
)

// RegisterModFn registers new modifier function.
func RegisterModFn(name, alias string, mod ModFn) *ModFnTuple {
	if idx, ok := modRegistry[name]; ok && idx >= 0 && idx < len(modBuf) {
		return &modBuf[idx]
	}
	modBuf = append(modBuf, ModFnTuple{
		docgen: docgen{
			name:  name,
			alias: alias,
		},
		fn: mod,
	})
	idx := len(modBuf) - 1
	modRegistry[name] = idx
	if len(alias) > 0 {
		modRegistry[alias] = idx
	}
	return &modBuf[idx]
}

// RegisterModFnNS registers new modifier function in given namespace.
func RegisterModFnNS(namespace, name, alias string, mod ModFn) *ModFnTuple {
	name = namespace + "::" + name
	if len(alias) > 0 {
		alias = namespace + "::" + alias
	}
	return RegisterModFn(name, alias, mod)
}

// GetModFn returns modifier from the registry.
func GetModFn(name string) ModFn {
	if idx, ok := modRegistry[name]; ok && idx >= 0 && idx < len(modBuf) {
		return modBuf[idx].fn
	}
	return nil
}
