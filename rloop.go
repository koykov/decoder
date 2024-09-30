package decoder

import (
	"github.com/koykov/byteconv"
	"github.com/koykov/inspector"
)

const (
	// RangeLoop object statuses.
	rlFree  = uint(0)
	rlInuse = uint(1)
)

// RangeLoop is a object that injects to inspector to perform range loop execution.
type RangeLoop struct {
	cntr  int
	stat  uint
	n     *node
	nodes []node
	ctx   *Ctx
	next  *RangeLoop
}

// NewRangeLoop makes new RL.
func NewRangeLoop(r *node, nodes []node, ctx *Ctx) *RangeLoop {
	rl := RangeLoop{
		n:     r,
		nodes: nodes,
		ctx:   ctx,
	}
	return &rl
}

// RequireKey checks if node requires a key to store in the context.
func (rl *RangeLoop) RequireKey() bool {
	return len(rl.n.loopKey) > 0
}

// SetKey saves key to the context.
func (rl *RangeLoop) SetKey(val any, ins inspector.Inspector) {
	rl.ctx.Set(byteconv.B2S(rl.n.loopKey), val, ins)
}

// SetVal saves value to the context.
func (rl *RangeLoop) SetVal(val any, ins inspector.Inspector) {
	rl.ctx.Set(byteconv.B2S(rl.n.loopVal), val, ins)
}

// Iterate performs the iteration.
func (rl *RangeLoop) Iterate() inspector.LoopCtl {
	if rl.ctx.brkD > 0 {
		return inspector.LoopCtlBrk
	}

	rl.cntr++
	var err, lerr error
	for i := 0; i < len(rl.n.child); i++ {
		ch := &rl.n.child[i]
		err = followRule(ch, rl.ctx)
		if err == ErrLBreakLoop {
			lerr = err
		}
		if err == ErrBreakLoop {
			if rl.ctx.brkD > 0 {
				rl.ctx.brkD--
			}
			return inspector.LoopCtlBrk
		}
		if err == ErrContLoop {
			return inspector.LoopCtlCnt
		}
	}
	if err == ErrBreakLoop || lerr == ErrLBreakLoop {
		if rl.ctx.brkD > 0 {
			rl.ctx.brkD--
		}
		return inspector.LoopCtlBrk
	}
	return inspector.LoopCtlNone
}

// Reset clears all data in the list of RL.
func (rl *RangeLoop) Reset() {
	crl := rl
	for crl != nil {
		crl.stat = rlFree
		crl.cntr = 0
		crl.ctx = nil
		crl.nodes = crl.nodes[:0]
		crl = crl.next
	}
}
