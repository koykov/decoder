package decoder

// Decoder represents main decoder object.
// Decoder contains only parsed ruleset.
// All temporary and intermediate data should be store in context logic to make using of decoders thread-safe.
type Decoder struct {
	ID   int
	Key  string
	tree *Tree
}

var decDB = initDB()

// RegisterDecoder saves decoder by ID and key in the registry.
//
// You may use to access to the decoder both ID or key.
// This function can be used in any time to register new decoders or overwrite existing to provide dynamics.
func RegisterDecoder(id int, key string, tree *Tree) {
	decDB.set(id, key, tree)
}

// RegisterDecoderID saves decoder using only ID.
//
// See RegisterDecoder().
func RegisterDecoderID(id int, tree *Tree) {
	decDB.set(id, "-1", tree)
}

// RegisterDecoderKey saves decoder using only key.
//
// See RegisterDecoder().
func RegisterDecoderKey(key string, tree *Tree) {
	decDB.set(-1, key, tree)
}

// Decode applies decoder rules using given id.
//
// ctx should contain all variables mentioned in the decoder's body.
func Decode(key string, ctx *Ctx) error {
	dec := decDB.getKey(key)
	if dec == nil {
		return ErrDecoderNotFound
	}
	// Decode corresponding ruleset.
	return DecodeRuleset(dec.tree.nodes, ctx)
}

// DecodeFallback applies decoder rules using one of keys: key or fallback key.
//
// Using this func you can handle cases when some objects have custom decoders and all other should use default decoders.
// Example:
// decoder registry:
// * decoderUser
// * decoderUser-15
// user object with id 15
// Call of decoder.DecoderFallback("decUser-15", "decUser", ctx) will take decoder decUser-15 from registry.
// In other case, for user #4:
// call of decoder.DecoderFallback("decUser-4", "decUser", ctx) will take default decoder decUser from registry.
func DecodeFallback(key, fbKey string, ctx *Ctx) error {
	dec := decDB.getKey1(key, fbKey)
	if dec == nil {
		return ErrDecoderNotFound
	}
	// Decode corresponding ruleset.
	return DecodeRuleset(dec.tree.nodes, ctx)
}

// DecodeByID applies decoder rules using given id.
func DecodeByID(id int, ctx *Ctx) error {
	dec := decDB.getID(id)
	if dec == nil {
		return ErrDecoderNotFound
	}
	// Decode corresponding ruleset.
	return DecodeRuleset(dec.tree.nodes, ctx)
}

// DecodeRuleset applies decoder ruleset without using id.
func DecodeRuleset(ruleset Ruleset, ctx *Ctx) (err error) {
	n := len(ruleset)
	if n == 0 {
		return
	}
	_ = ruleset[n-1]
	for i := 0; i < n; i++ {
		if err = followRule(&ruleset[i], ctx); err != nil {
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
	case r.typ == typeCond:
		// todo implement me
	case r.typ == typeCondOK:
		// todo implement me
	case r.typ == typeCondTrue || r.typ == typeCondFalse:
		// todo implement me
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

var _, _, _, _ = RegisterDecoder, RegisterDecoderID, DecodeFallback, DecodeByID
