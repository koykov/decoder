package decoder

// Signature of callback function.
//
// args contains list of all arguments you passed in decoder rule.
type CallbackFn func(ctx *Ctx, args []interface{}) error

var (
	// Registry of callback functions.
	callbackRegistry = map[string]CallbackFn{}
)

// Add new callback to the registry.
func RegisterCallbackFn(name, alias string, cb CallbackFn) {
	callbackRegistry[name] = cb
	if len(alias) > 0 {
		callbackRegistry[alias] = cb
	}
}

// Get callback function from the registry.
func GetCallbackFn(name string) *CallbackFn {
	if fn, ok := callbackRegistry[name]; ok {
		return &fn
	}
	return nil
}
