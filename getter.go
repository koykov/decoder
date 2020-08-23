package decoder

type GetterFn func(ctx *Ctx, buf *interface{}, args []interface{}) error

var (
	getterRegistry = map[string]GetterFn{}
)

func RegisterGetterFn(name, alias string, cb GetterFn) {
	getterRegistry[name] = cb
	if len(alias) > 0 {
		getterRegistry[alias] = cb
	}
}

func GetGetterFn(name string) *GetterFn {
	if fn, ok := getterRegistry[name]; ok {
		return &fn
	}
	return nil
}
