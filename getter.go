package decoder

// GetterFn represents signature of getter callback function.
//
// args contains list of all arguments you passed in decoder rule.
type GetterFn func(ctx *Ctx, buf *any, args []any) error

type GetterFnTuple struct {
	docgen
	fn GetterFn
}

var (
	// Registry of getter callback functions.
	getterRegistry = map[string]int{}
	getterBuf      []GetterFnTuple

	_ = RegisterGetterFnNS
)

// RegisterGetterFn registers new getter callback to the registry.
func RegisterGetterFn(name, alias string, cb GetterFn) *GetterFnTuple {
	if idx, ok := getterRegistry[alias]; ok && idx >= 0 && idx < len(getterBuf) {
		return &getterBuf[idx]
	}
	getterBuf = append(getterBuf, GetterFnTuple{
		docgen: docgen{
			name:  name,
			alias: alias,
		},
		fn: cb,
	})
	idx := len(getterBuf) - 1
	getterRegistry[name] = idx
	if len(alias) > 0 {
		getterRegistry[alias] = idx
	}
	return &getterBuf[idx]
}

// RegisterGetterFnNS registers new getter callback in given namespace.
func RegisterGetterFnNS(namespace, name, alias string, cb GetterFn) *GetterFnTuple {
	name = namespace + "::" + name
	if len(alias) > 0 {
		alias = namespace + "::" + alias
	}
	return RegisterGetterFn(name, alias, cb)
}

// GetGetterFn returns getter callback function from the registry.
func GetGetterFn(name string) GetterFn {
	if idx, ok := getterRegistry[name]; ok && idx >= 0 && idx < len(getterBuf) {
		return getterBuf[idx].fn
	}
	return nil
}
