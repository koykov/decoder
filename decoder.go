package decoder

import (
	"sync"
)

// Main decoder object.
// Decoder contains only parsed rules.
// All temporary and intermediate data should be store in context logic to make using of decoders thread-safe.
type Decoder struct {
	Id    string
	rules Rules
}

var (
	// Decoders registry.
	mux             sync.Mutex
	decoderRegistry = map[string]*Decoder{}
)

// Register decoder rules in the registry.
func RegisterDecoder(id string, rules Rules) {
	decoder := Decoder{
		Id:    id,
		rules: rules,
	}
	mux.Lock()
	decoderRegistry[id] = &decoder
	mux.Unlock()
}

// Apply decoder rules using given id.
//
// ctx should contains all variables mentioned in the decoder's body.
func Decode(id string, ctx *Ctx) error {
	var (
		decoder *Decoder
		ok      bool
	)
	mux.Lock()
	decoder, ok = decoderRegistry[id]
	mux.Unlock()
	if !ok {
		return ErrDecoderNotFound
	}
	// Decode corresponding ruleset.
	return DecodeRules(decoder.rules, ctx)
}

// Apply decoder rules without using id.
func DecodeRules(rules Rules, ctx *Ctx) (err error) {
	for _, rule := range rules {
		err = followRule(&rule, ctx)
		if err != nil {
			return
		}
	}
	return
}

// Generic function to apply single rule.
func followRule(rule *rule, ctx *Ctx) (err error) {
	switch {
	case rule.callback != nil:
		// Rule is a callback.
		// Collect arguments.
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
		// Execute callback func.
		err = (*rule.callback)(ctx, ctx.bufA)
	case rule.getter != nil:
		// F2V rule.
		// Collect arguments.
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
		// Call getter callback func.
		err = (*rule.getter)(ctx, &ctx.bufX, ctx.bufA)
		if err != nil {
			return
		}
		// Assign result to destination.
		err = ctx.set(rule.dst, ctx.bufX, rule.ins)
	case len(rule.dst) > 0 && len(rule.src) > 0 && rule.static:
		// V2V rule with static source.
		// Just assign the source it to destination.
		ctx.buf = append(ctx.buf[:0], rule.src...)
		err = ctx.set(rule.dst, &ctx.buf, rule.ins)
	case len(rule.dst) > 0 && len(rule.src) > 0 && !rule.static:
		// V2V rule with dynamic source.
		// Get source value.
		raw := ctx.get(rule.src, rule.subset)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		// Apply modifiers.
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
		// Assign to destination.
		err = ctx.set(rule.dst, raw, rule.ins)
	}
	return
}
