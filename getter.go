package decoder

// GetterFn represents signature of getter callback function.
//
// args contains list of all arguments you passed in decoder rule.
type GetterFn func(ctx *Ctx, buf *any, args []any) error

var (
	// Registry of getter callback functions.
	getterRegistry = map[string]GetterFn{}
)

// RegisterGetterFn registers new getter callback to the registry.
func RegisterGetterFn(name, alias string, cb GetterFn) {
	getterRegistry[name] = cb
	if len(alias) > 0 {
		getterRegistry[alias] = cb
	}
}

// GetGetterFn returns getter callback function from the registry.
func GetGetterFn(name string) *GetterFn {
	if fn, ok := getterRegistry[name]; ok {
		return &fn
	}
	return nil
}
