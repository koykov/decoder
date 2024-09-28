package decoder

import (
	"sync"
)

// Decoder represents main decoder object.
// Decoder contains only parsed ruleset.
// All temporary and intermediate data should be store in context logic to make using of decoders thread-safe.
type Decoder struct {
	Id string
	rs Tree
}

var (
	// Decoders registry.
	mux      sync.Mutex
	registry = map[string]*Decoder{}
)

// RegisterDecoder registers decoder ruleset in the registry.
func RegisterDecoder(id string, rules Tree) {
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
func DecodeRuleset(ruleset Tree, ctx *Ctx) (err error) {
	n := len(ruleset.nodes)
	if n == 0 {
		return nil
	}
	_ = ruleset.nodes[n-1]
	for i := 0; i < n; i++ {
		if err = followRule(&ruleset.nodes[i], ctx); err != nil {
			return
		}
	}
	return
}

// Generic function to apply single node.
func followRule(r *node, ctx *Ctx) (err error) {
	switch {
	case r.typ == typeLoopRange:
		// Evaluate range loops.
		// See Ctx.rloop().
		ctx.brkD = 0
		ctx.rloop(r.loopSrc, r, r.child)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
	case r.typ == typeLoopCount:
		// Evaluate counter loops.
		// See Ctx.cloop().
		ctx.brkD = 0
		ctx.cloop(r, r.child)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
	case r.typ == typeBreak:
		// todo cover with test after condition implementation
		// Break the loop.
		ctx.brkD = r.loopBrkD
		err = ErrBreakLoop
	case r.typ == typeLBreak:
		// todo cover with test after condition implementation
		// Lazy break the loop.
		ctx.brkD = r.loopBrkD
		err = ErrLBreakLoop
	case r.typ == typeContinue:
		// todo cover with test after condition implementation
		// Go to next iteration of loop.
		err = ErrContLoop
	case r.callback != nil:
		// Rule is a callback.
		// Collect arguments.
		ctx.bufA = ctx.bufA[:0]
		if n := len(r.arg); n > 0 {
			_ = r.arg[n-1]
			for i := 0; i < n; i++ {
				a := r.arg[i]
				if a.static {
					ctx.bufA = append(ctx.bufA, &a.val)
				} else {
					val := ctx.get(a.val, a.subset)
					ctx.bufA = append(ctx.bufA, val)
				}
			}
		}
		// Execute callback func.
		err = r.callback(ctx, ctx.bufA)
	case r.getter != nil:
		// F2V node.
		// Collect arguments.
		ctx.bufA = ctx.bufA[:0]
		if n := len(r.arg); n > 0 {
			_ = r.arg[n-1]
			for i := 0; i < n; i++ {
				a := r.arg[i]
				if a.static {
					ctx.bufA = append(ctx.bufA, &a.val)
				} else {
					val := ctx.get(a.val, a.subset)
					ctx.bufA = append(ctx.bufA, val)
				}
			}
		}
		// Call getter callback func.
		err = r.getter(ctx, &ctx.bufX, ctx.bufA)
		if err != nil {
			return
		}
		// Assign result to destination.
		err = ctx.set(r.dst, ctx.bufX, r.ins)
	case len(r.dst) > 0 && len(r.src) > 0 && r.static:
		// V2V node with static source.
		// Just assign the source it to destination.
		ctx.buf = append(ctx.buf[:0], r.src...)
		err = ctx.set(r.dst, &ctx.buf, r.ins)
	case len(r.dst) > 0 && len(r.src) > 0 && !r.static:
		// V2V node with dynamic source.
		// Get source value.
		raw := ctx.get(r.src, r.subset)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		// Apply modifiers.
		if n := len(r.mod); n > 0 {
			_ = r.mod[n-1]
			for i := 0; i < n; i++ {
				m := &r.mod[i]
				// Collect arguments to buffer.
				ctx.bufA = ctx.bufA[:0]
				if k := len(m.arg); k > 0 {
					_ = m.arg[k-1]
					for j := 0; j < k; j++ {
						a := m.arg[j]
						if a.static {
							ctx.bufA = append(ctx.bufA, &a.val)
						} else {
							val := ctx.get(a.val, a.subset)
							ctx.bufA = append(ctx.bufA, val)
						}
					}
				}
				ctx.bufX = raw
				// Call the modifier func.
				ctx.Err = m.fn(ctx, &ctx.bufX, ctx.bufX, ctx.bufA)
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
		err = ctx.set(r.dst, raw, r.ins)
	}
	return
}
