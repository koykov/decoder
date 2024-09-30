package decoder

// CallbackFn represents the signature of callback function.
//
// args contains list of all arguments you passed in decoder node.
type CallbackFn func(ctx *Ctx, args []any) error

type CallbackFnTuple struct {
	docgen
	fn CallbackFn
}

var (
	// Registry of callback functions.
	callbackRegistry = map[string]int{}
	callbackBuf      []CallbackFnTuple

	_, _ = RegisterCallbackFn, RegisterCallbackFnNS
)

// RegisterCallbackFn registers new callback to the registry.
func RegisterCallbackFn(name, alias string, cb CallbackFn) *CallbackFnTuple {
	if idx, ok := callbackRegistry[name]; ok && idx >= 0 && idx < len(callbackBuf) {
		return &callbackBuf[idx]
	}
	callbackBuf = append(callbackBuf, CallbackFnTuple{
		docgen: docgen{
			name:  name,
			alias: alias,
		},
		fn: cb,
	})
	idx := len(callbackBuf) - 1
	callbackRegistry[name] = idx
	if len(alias) > 0 {
		callbackRegistry[alias] = idx
	}
	return &callbackBuf[idx]
}

// RegisterCallbackFnNS registers new callback in given namespace.
func RegisterCallbackFnNS(namespace, name, alias string, cb CallbackFn) *CallbackFnTuple {
	name = namespace + "::" + name
	if len(alias) > 0 {
		alias = namespace + "::" + alias
	}
	return RegisterCallbackFn(name, alias, cb)
}

// GetCallbackFn returns callback function from the registry.
func GetCallbackFn(name string) CallbackFn {
	if idx, ok := callbackRegistry[name]; ok && idx >= 0 && idx < len(callbackBuf) {
		return callbackBuf[idx].fn
	}
	return nil
}
