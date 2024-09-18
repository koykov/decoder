package decoder

import (
	"strconv"

	"github.com/koykov/byteconv"
)

// Counter loop method to evaluate expressions like:
// for i:=0; i<10; i++ { ... }
func (ctx *Ctx) cloop(r *rule, rs Ruleset) {
	var (
		cnt, lim  int64
		allowIter bool
	)
	// Prepare bounds.
	cnt = ctx.cloopRange(r.loopCntStatic, r.loopCntInit)
	if ctx.Err != nil {
		return
	}
	lim = ctx.cloopRange(r.loopLimStatic, r.loopLim)
	if ctx.Err != nil {
		return
	}
	// Prepare counters.
	ctx.bufLC = append(ctx.bufLC, cnt)
	idxLC := len(ctx.bufLC) - 1
	valLC := cnt
	// Start the loop.
	allowIter = false
	for {
		// Check iteration allowance.
		switch r.loopCondOp {
		case OpLt:
			allowIter = valLC < lim
		case OpLtq:
			allowIter = valLC <= lim
		case OpGt:
			allowIter = valLC > lim
		case OpGtq:
			allowIter = valLC >= lim
		case OpEq:
			allowIter = valLC == lim
		case OpNq:
			allowIter = valLC != lim
		default:
			ctx.Err = ErrWrongLoopCond
			break
		}
		// Check breakN signal from child loop.
		allowIter = allowIter && ctx.brkD == 0

		if !allowIter {
			break
		}

		// Set/update counter var.
		ctx.SetStatic(byteconv.B2S(r.loopCnt), &ctx.bufLC[idxLC])

		// Loop over child nodes with square brackets check in paths.
		ctx.chQB = true
		var err, lerr error
		for i := 0; i < len(r.child); i++ {
			ch := &r.child[i]
			err = followRule(ch, ctx)
			if err == ErrLBreakLoop {
				lerr = err
			}
			if err == ErrBreakLoop || err == ErrContLoop {
				break
			}
		}
		ctx.chQB = false

		// Modify counter var.
		switch r.loopCntOp {
		case OpInc:
			valLC++
			ctx.bufLC[idxLC]++
		case OpDec:
			valLC--
			ctx.bufLC[idxLC]--
		default:
			ctx.Err = ErrWrongLoopOp
			break
		}

		// Handle break/continue cases.
		if err == ErrBreakLoop || lerr == ErrLBreakLoop {
			if ctx.brkD > 0 {
				ctx.brkD--
			}
			break
		}
		if err == ErrContLoop {
			continue
		}
	}
	return
}

// Counter loop bound check helper.
//
// Converts initial and final values of the counter to static int values.
func (ctx *Ctx) cloopRange(static bool, b []byte) (r int64) {
	if static {
		r, ctx.Err = strconv.ParseInt(byteconv.B2S(b), 0, 0)
		if ctx.Err != nil {
			return
		}
	} else {
		var ok bool
		raw := ctx.get(b, nil)
		if ctx.Err != nil {
			return
		}
		r, ok = iface2int(raw)
		if !ok {
			ctx.Err = ErrWrongLoopLim
			return
		}
	}
	return
}
