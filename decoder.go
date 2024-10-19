package decoder

import (
	"github.com/koykov/bytealg"
	"github.com/koykov/byteconv"
	"github.com/koykov/inspector"
)

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
	case r.typ == typeCondOK:
		// Condition-OK node evaluates expressions like if-ok with helper.
		var ok bool
		// Check condition-OK helper (mandatory at all).
		if len(r.condHlp) > 0 {
			fn := GetCondOKFn(byteconv.B2S(r.condHlp))
			if fn == nil {
				err = ErrCondHlpNotFound
				return
			}
			// Prepare arguments list.
			ctx.bufA = ctx.bufA[:0]
			if n := len(r.condHlpArg); n > 0 {
				_ = r.condHlpArg[n-1]
				for i := 0; i < n; i++ {
					arg_ := r.condHlpArg[i]
					if arg_.static {
						ctx.bufA = append(ctx.bufA, &arg_.val)
					} else {
						val := ctx.get(arg_.val, arg_.subset)
						ctx.bufA = append(ctx.bufA, val)
					}
				}
			}
			// Call condition-ok helper func.
			fn(ctx, &ctx.bufX, &ctx.bufBl, ctx.bufA)
			ok = ctx.bufBl
			// Set var, ok to context.
			lv, lr := byteconv.B2S(r.condOKL), byteconv.B2S(r.condOKR)
			insn := byteconv.B2S(r.condIns)
			if len(insn) == 0 {
				insn = "static"
			}
			ins, err := inspector.GetInspector(insn)
			if err != nil {
				return err
			}
			raw := ctx.bufX
			ctx.Set(lv, raw, ins)
			ctx.SetStatic(lr, ctx.bufBl)

			// Check extended condition (eg: !ok).
			if len(r.condR) > 0 {
				ok, err = nodeCmp(r, ctx)
			}
			// Evaluate condition.
			if ok {
				// True case.
				if len(r.child) > 0 {
					err = followRule(&r.child[0], ctx)
				}
			} else {
				// Else case.
				if len(r.child) > 1 {
					err = followRule(&r.child[1], ctx)
				}
			}
		}
	case r.typ == typeCond:
		// Condition node evaluates condition expressions.
		var ok bool
		switch {
		case len(r.condHlp) > 0 && r.condLC == lcNone:
			// Condition helper caught (no LC case).
			fn := GetCondFn(byteconv.B2S(r.condHlp))
			if fn == nil {
				err = ErrCondHlpNotFound
				return
			}
			// Prepare arguments list.
			ctx.bufA = ctx.bufA[:0]
			if n := len(r.condHlpArg); n > 0 {
				_ = r.condHlpArg[n-1]
				for i := 0; i < len(r.condHlpArg); i++ {
					arg_ := r.condHlpArg[i]
					if arg_.static {
						ctx.bufA = append(ctx.bufA, &arg_.val)
					} else {
						val := ctx.get(arg_.val, arg_.subset)
						ctx.bufA = append(ctx.bufA, val)
					}
				}
			}
			// Call condition helper func.
			ok = fn(ctx, ctx.bufA)
		case len(r.condHlp) > 0 && r.condLC > lcNone:
			// Condition helper in LC mode.
			if len(r.condHlpArg) == 0 {
				err = ErrModNoArgs
				return
			}
			ok = ctx.cmpLC(r.condLC, r.condHlpArg[0].val, r.condOp, r.condR)
		default:
			ok, err = nodeCmp(r, ctx)
		}
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		// Evaluate condition.
		if ok {
			// True case.
			if len(r.child) > 0 {
				err = followRule(&r.child[0], ctx)
			}
		} else {
			// Else case.
			if len(r.child) > 1 {
				err = followRule(&r.child[1], ctx)
			}
		}
	case r.typ == typeCondTrue || r.typ == typeCondFalse:
		if err = DecodeRuleset(r.child, ctx); err != nil {
			return
		}
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
		// F2V r.
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

func (ctx *Ctx) cmpLC(lc lc, path []byte, cond op, right []byte) bool {
	ctx.Err = nil
	if ctx.chQB {
		path = ctx.replaceQB(path)
	}

	ctx.bufS = ctx.bufS[:0]
	ctx.bufS = bytealg.AppendSplitString(ctx.bufS, byteconv.B2S(path), ".", -1)
	if len(ctx.bufS) == 0 {
		return false
	}

	for i := 0; i < ctx.ln; i++ {
		v := &ctx.vars[i]
		if v.key == ctx.bufS[0] {
			switch lc {
			case lcLen:
				ctx.Err = v.ins.Length(v.val, &ctx.bufI_, ctx.bufS[1:]...)
			case lcCap:
				ctx.Err = v.ins.Capacity(v.val, &ctx.bufI_, ctx.bufS[1:]...)
			default:
				return false
			}
			if ctx.Err != nil {
				return false
			}
			si := inspector.StaticInspector{}
			ctx.bufBl = false
			ctx.Err = si.Compare(ctx.bufI_, inspector.Op(cond), byteconv.B2S(right), &ctx.bufBl)
			return ctx.bufBl
		}
	}
	return false
}

// Evaluate condition expressions.
func nodeCmp(node *node, ctx *Ctx) (r bool, err error) {
	// Regular comparison.
	sl := node.condStaticL
	sr := node.condStaticR
	if sl && sr {
		// It's senseless to compare two static values.
		err = ErrSenselessCond
		return
	}
	if sr {
		// Right side is static. This is preferred case
		r = ctx.cmp(node.condL, node.condOp, node.condR)
	} else if sl {
		// Left side is static.
		// dyntpl can't handle expressions like {% if 10 > item.Weight %}...
		// therefore it inverts condition to {% if item.Weight < 10 %}...
		r = ctx.cmp(node.condR, node.condOp.Swap(), node.condL)
	} else {
		// Both sides aren't static. This is a bad case, since need to inspect variables twice.
		ctx.get(node.condR, nil)
		if ctx.Err == nil {
			if err = ctx.BufAcc.StakeOut().WriteX(ctx.bufX).Error(); err != nil {
				return
			}
			r = ctx.cmp(node.condL, node.condOp, ctx.BufAcc.StakedBytes())
		}
	}
	return
}

var _, _, _, _ = RegisterDecoder, RegisterDecoderID, DecodeFallback, DecodeByID
