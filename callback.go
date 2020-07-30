package jsondecoder

type CallbackFn func(ctx *Ctx, args []interface{}) error

var (
	callbackRegistry = map[string]CallbackFn{}
)

func RegisterCallbackFn(name, alias string, cb CallbackFn) {
	callbackRegistry[name] = cb
	if len(alias) > 0 {
		callbackRegistry[alias] = cb
	}
}

func GetCallbackFn(name string) *CallbackFn {
	if fn, ok := callbackRegistry[name]; ok {
		return &fn
	}
	return nil
}
