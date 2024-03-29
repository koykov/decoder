package decoder

import (
	"sync"
)

// Decoder represents main decoder object.
// Decoder contains only parsed ruleset.
// All temporary and intermediate data should be store in context logic to make using of decoders thread-safe.
type Decoder struct {
	Id string
	rs Ruleset
}

var (
	// Decoders registry.
	mux      sync.Mutex
	registry = map[string]*Decoder{}
)

// RegisterDecoder registers decoder ruleset in the registry.
func RegisterDecoder(id string, rules Ruleset) {
	decoder := Decoder{
		Id: id,
		rs: rules,
	}
	mux.Lock()
	registry[id] = &decoder
	mux.Unlock()
}

// Decode applies decoder rules using given id.
//
// ctx should contain all variables mentioned in the decoder's body.
func Decode(id string, ctx *Ctx) error {
	var (
		decoder *Decoder
		ok      bool
	)
	mux.Lock()
	decoder, ok = registry[id]
	mux.Unlock()
	if !ok {
		return ErrDecoderNotFound
	}
	// Decode corresponding ruleset.
	return DecodeRuleset(decoder.rs, ctx)
}

// DecodeRuleset applies decoder ruleset without using id.
func DecodeRuleset(ruleset Ruleset, ctx *Ctx) (err error) {
	for _, rule := range ruleset {
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
			for _, arg_ := range rule.arg {
				if arg_.static {
					ctx.bufA = append(ctx.bufA, &arg_.val)
				} else {
					val := ctx.get(arg_.val, arg_.subset)
					ctx.bufA = append(ctx.bufA, val)
				}
			}
		}
		// Execute callback func.
		err = rule.callback(ctx, ctx.bufA)
	case rule.getter != nil:
		// F2V rule.
		// Collect arguments.
		ctx.bufA = ctx.bufA[:0]
		if len(rule.arg) > 0 {
			for _, arg_ := range rule.arg {
				if arg_.static {
					ctx.bufA = append(ctx.bufA, &arg_.val)
				} else {
					val := ctx.get(arg_.val, arg_.subset)
					ctx.bufA = append(ctx.bufA, val)
				}
			}
		}
		// Call getter callback func.
		err = rule.getter(ctx, &ctx.bufX, ctx.bufA)
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
			for _, mod_ := range rule.mod {
				// Collect arguments to buffer.
				ctx.bufA = ctx.bufA[:0]
				if len(mod_.arg) > 0 {
					for _, arg_ := range mod_.arg {
						if arg_.static {
							ctx.bufA = append(ctx.bufA, &arg_.val)
						} else {
							val := ctx.get(arg_.val, arg_.subset)
							ctx.bufA = append(ctx.bufA, val)
						}
					}
				}
				ctx.bufX = raw
				// Call the modifier func.
				ctx.Err = mod_.fn(ctx, &ctx.bufX, ctx.bufX, ctx.bufA)
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
