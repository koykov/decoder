package decoder

import "sync"

type Decoder struct {
	Id    string
	rules Rules
}

var (
	mux             sync.Mutex
	decoderRegistry = map[string]*Decoder{}
)

func RegisterDecoder(id string, rules Rules) {
	decoder := Decoder{
		Id:    id,
		rules: rules,
	}
	mux.Lock()
	decoderRegistry[id] = &decoder
	mux.Unlock()
}

func Decode(id string, ctx *Ctx) error {
	mux.Lock()
	decoder, ok := decoderRegistry[id]
	mux.Unlock()
	if !ok {
		return ErrDecoderNotFound
	}
	return DecodeRules(decoder.rules, ctx)
}

func DecodeRules(rules Rules, ctx *Ctx) (err error) {
	for _, rule := range rules {
		err = followRule(&rule, ctx)
		if err != nil {
			return
		}
	}
	return
}

func followRule(rule *rule, ctx *Ctx) (err error) {
	switch {
	case rule.callback != nil:
		ctx.bufA = ctx.bufA[:0]
		if len(rule.arg) > 0 {
			for _, arg := range rule.arg {
				if arg.static {
					ctx.bufA = append(ctx.bufA, &arg.val)
				} else {
					val := ctx.get(arg.val, arg.subset)
					ctx.bufA = append(ctx.bufA, val)
				}
			}
		}
		err = (*rule.callback)(ctx, ctx.bufA)
	case rule.getter != nil:
		ctx.bufA = ctx.bufA[:0]
		if len(rule.arg) > 0 {
			for _, arg := range rule.arg {
				if arg.static {
					ctx.bufA = append(ctx.bufA, &arg.val)
				} else {
					val := ctx.get(arg.val, arg.subset)
					ctx.bufA = append(ctx.bufA, val)
				}
			}
		}
		err = (*rule.getter)(ctx, &ctx.bufX, ctx.bufA)
		if err != nil {
			return
		}
		err = ctx.set(rule.dst, ctx.bufX)
	case len(rule.dst) > 0 && len(rule.src) > 0 && rule.static:
		ctx.buf = append(ctx.buf[:0], rule.src...)
		err = ctx.set(rule.dst, &ctx.buf)
	case len(rule.dst) > 0 && len(rule.src) > 0 && !rule.static:
		raw := ctx.get(rule.src, rule.subset)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		if len(rule.mod) > 0 {
			for _, mod := range rule.mod {
				// Collect arguments to buffer.
				ctx.bufA = ctx.bufA[:0]
				if len(mod.arg) > 0 {
					for _, arg := range mod.arg {
						if arg.static {
							ctx.bufA = append(ctx.bufA, &arg.val)
						} else {
							val := ctx.get(arg.val, arg.subset)
							ctx.bufA = append(ctx.bufA, val)
						}
					}
				}
				ctx.bufX = raw
				// Call the modifier func.
				ctx.Err = (*mod.fn)(ctx, &ctx.bufX, ctx.bufX, ctx.bufA)
				if ctx.Err != nil {
					break
				}
				raw = ctx.bufX
			}
		}
		if ctx.Err != nil {
			return
		}
		err = ctx.set(rule.dst, raw)
	}
	return
}
