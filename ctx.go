package decoder

import (
	"bytes"
	"strings"

	"github.com/koykov/bytealg"
	"github.com/koykov/bytebuf"
	"github.com/koykov/byteconv"
	"github.com/koykov/inspector"
	"github.com/koykov/vector"
	"github.com/koykov/vector_inspector"
)

// Ctx represents decoder context object.
//
// Contains list of variables that can be used as source or destination.
type Ctx struct {
	// List of context variables and list len.
	vars []ctxVar
	ln   int

	// Check square brackets flag.
	chQB bool

	// Internal buffers.
	accB  []byte
	buf   []byte
	bufBB [][]byte
	lenBB int
	bufS  []string
	bufI  int64
	bufI_ int
	bufU  uint64
	bufF  float64
	bufBl bool
	bufX  any
	bufA  []any
	bufLC []int64
	// Range loop helper.
	rl *RangeLoop

	// Break depth.
	brkD int

	// List of variables taken from ipools and registered to return back.
	ipv  []ipoolVar
	ipvl int

	// External buffers to use in modifier and condition helpers.
	BufAcc bytebuf.Accumulative
	// todo remove as unused later
	Buf, Buf1, Buf2 bytebuf.Chain

	Err error
}

// Context variable object.
type ctxVar struct {
	key string
	val any
	ins inspector.Inspector
}

var (
	qbL = []byte("[")
	qbR = []byte("]")
)

// NewCtx makes new context object.
func NewCtx() *Ctx {
	ctx := Ctx{
		vars: make([]ctxVar, 0),
		bufS: make([]string, 0),
		bufA: make([]any, 0),
	}
	return &ctx
}

// Set the variable to context.
// Inspector ins should be corresponded to variable val.
func (ctx *Ctx) Set(key string, val any, ins inspector.Inspector) {
	for i := 0; i < ctx.ln; i++ {
		if ctx.vars[i].key == key {
			// Update existing variable.
			ctx.vars[i].val = val
			ctx.vars[i].ins = ins
			return
		}
	}
	// Add new variable.
	if ctx.ln < len(ctx.vars) {
		// Use existing item in variable list.
		ctx.vars[ctx.ln].key = key
		ctx.vars[ctx.ln].val = val
		ctx.vars[ctx.ln].ins = ins
	} else {
		// Extend the variable list with new one.
		ctx.vars = append(ctx.vars, ctxVar{
			key: key,
			val: val,
			ins: ins,
		})
	}
	// Increase variables count.
	ctx.ln++
}

// SetStatic registers static variable in context.
func (ctx *Ctx) SetStatic(key string, val any) {
	ins, err := inspector.GetInspector("static")
	if err != nil {
		ctx.Err = err
		return
	}
	ctx.Set(key, val, ins)
}

// SetVector directly register vector in context.
func (ctx *Ctx) SetVector(key string, vec vector.Interface) {
	if vec == nil {
		return
	}
	ctx.Set(key, vec, vector_inspector.VectorInspector{})
}

// SetVectorNode directly registers vector's node in context under given key.
func (ctx *Ctx) SetVectorNode(key string, node *vector.Node) error {
	if node == nil || node.Type() == vector.TypeNull {
		return ErrEmptyNode
	}
	ctx.Set(key, node, vector_inspector.VectorInspector{})
	return nil
}

// Get arbitrary value from the context by path.
//
// See Ctx.get().
// Path syntax: <ctxVrName>[.<Field>[.<NestedField0>[....<NestedFieldN>]]]
// Examples:
// * user.Bio.Birthday
// * staticVar
func (ctx *Ctx) Get(path string) any {
	return ctx.get(byteconv.S2B(path), nil)
}

// AcquireBytes returns accumulative buffer.
func (ctx *Ctx) AcquireBytes() []byte {
	return ctx.accB
}

// ReleaseBytes updates accumulative buffer with p.
func (ctx *Ctx) ReleaseBytes(p []byte) {
	if len(p) == 0 {
		return
	}
	ctx.accB = p
}

func (ctx *Ctx) Bufferize(p []byte) []byte {
	off := len(ctx.accB)
	ctx.accB = append(ctx.accB, p...)
	return ctx.accB[off:]
}

func (ctx *Ctx) BufferizeString(s string) string {
	off := len(ctx.accB)
	ctx.accB = append(ctx.accB, s...)
	return byteconv.B2S(ctx.accB[off:])
}

// AcquireFrom receives new variable from given pool and register it to return batch after finish decoder processing.
func (ctx *Ctx) AcquireFrom(pool string) (any, error) {
	v, err := ipoolRegistry.acquire(pool)
	if err != nil {
		return nil, err
	}
	if ctx.ipvl < len(ctx.ipv) {
		ctx.ipv[ctx.ipvl].key = pool
		ctx.ipv[ctx.ipvl].val = v
	} else {
		ctx.ipv = append(ctx.ipv, ipoolVar{key: pool, val: v})
	}
	ctx.ipvl++
	return v, nil
}

func (ctx *Ctx) reserveBB() int {
	if len(ctx.bufBB) == ctx.lenBB {
		ctx.bufBB = append(ctx.bufBB, nil)
	}
	ctx.lenBB++
	return ctx.lenBB - 1
}

// Internal getter.
func (ctx *Ctx) get(path []byte, subset [][]byte) any {
	if len(path) == 0 || ctx.ln == 0 {
		return nil
	}

	// Split path to separate words using dot as separator.
	ctx.splitPath(byteconv.B2S(path), ".")
	if len(ctx.bufS) == 0 {
		return nil
	}

	// Look for first path chunk in vars.
	_ = ctx.vars[ctx.ln-1]
	for i := 0; i < ctx.ln; i++ {
		v := &ctx.vars[i]
		if v.key == ctx.bufS[0] {
			// Var found.
			// Check var is vector or node.
			var (
				node *vector.Node
				ok   bool
			)
			switch x := v.val.(type) {
			case vector.Interface:
				node = x.Root()
				ok = true
			case *vector.Vector:
				node = x.Root()
				ok = true
			case *vector.Node:
				node = x
				ok = true
			}
			if ok && node != nil {
				// Var is vector node.
				if n := len(subset); n > 0 {
					// List of subsets provided.
					// Preserve item in []str buffer to check each key separately.
					ctx.bufS = append(ctx.bufS, "")
					_ = subset[n-1]
					for j := 0; j < n; j++ {
						if tail := subset[j]; len(tail) > 0 {
							// Fill preserved item with subset's value.
							ctx.bufS[len(ctx.bufS)-1] = byteconv.B2S(tail)
							ctx.bufX = node.Get(ctx.bufS[1:]...)
							if cn, ok := ctx.bufX.(*vector.Node); ok && cn.Type() != vector.TypeNull {
								// Successful hunt.
								break
							}
						}
					}
				} else {
					ctx.bufX = node.Get(ctx.bufS[1:]...)
				}
				return ctx.bufX
			}
			if v.ins != nil {
				// Variable is covered by inspector.
				ctx.Err = v.ins.GetTo(v.val, &ctx.bufX, ctx.bufS[1:]...)
				if ctx.Err != nil {
					return nil
				}
				return ctx.bufX
			}
			return v.val
		}
	}
	return nil
}

// Internal setter.
//
// Set val to destination by address path.
func (ctx *Ctx) set(path []byte, val any, insName []byte) error {
	if len(path) == 0 || ctx.ln == 0 {
		return nil
	}
	ctx.bufS = ctx.bufS[:0]
	ctx.bufS = bytealg.AppendSplit(ctx.bufS, byteconv.B2S(path), ".", -1)
	if len(ctx.bufS) == 0 {
		return nil
	}
	if ctx.bufS[0] == "ctx" || ctx.bufS[0] == "context" {
		if len(ctx.bufS) == 1 {
			// Attempt to overwrite the whole context object caught.
			return nil
		}
		// Var-to-ctx case.
		ctxPath := byteconv.B2S(path[len(ctx.bufS[0])+1:])
		if len(insName) > 0 {
			ins, err := inspector.GetInspector(byteconv.B2S(insName))
			if err != nil {
				return err
			}
			ctx.Set(ctxPath, val, ins)
		} else if node, ok := val.(*vector.Node); ok {
			_ = ctx.SetVectorNode(ctxPath, node)
		} else {
			ctx.SetStatic(ctxPath, val)
		}
		return nil
	}
	// Var-to-var case.
	_ = ctx.vars[ctx.ln-1]
	for i := 0; i < ctx.ln; i++ {
		v := &ctx.vars[i]
		if v.key == ctx.bufS[0] {
			if v.ins != nil {
				ctx.bufX = val
				ctx.Err = v.ins.SetWithBuffer(v.val, ctx.bufX, ctx, ctx.bufS[1:]...)
				if ctx.Err != nil {
					return ctx.Err
				}
			}
			break
		}
	}
	return nil
}

// Split path to separate words using dot as separator.
// So, path user.Bio.Birthday will convert to []string{"user", "Bio", "Birthday"}
func (ctx *Ctx) splitPath(path, separator string) {
	ctx.bufS = bytealg.AppendSplit(ctx.bufS[:0], path, separator, -1)
	ti := len(ctx.bufS) - 1
	if ti < 0 {
		return
	}
	tail := ctx.bufS[ti]
	if p := strings.IndexByte(tail, '@'); p != -1 {
		if p > 0 {
			if len(tail[p:]) > 1 {
				ctx.bufS = append(ctx.bufS, tail[p:])
			}
			ctx.bufS[ti] = ctx.bufS[ti][:p]
		}
	}
}

func (ctx *Ctx) rloop(path []byte, r *node, nodes []node) {
	ctx.bufS = ctx.bufS[:0]
	ctx.bufS = bytealg.AppendSplitString(ctx.bufS, byteconv.B2S(path), ".", -1)
	if len(ctx.bufS) == 0 {
		return
	}
	for i := 0; i < ctx.ln; i++ {
		v := &ctx.vars[i]
		if v.key == ctx.bufS[0] {
			// Look for free-range loop object in single-ordered list, see RangeLoop.
			var rl *RangeLoop
			if ctx.rl == nil {
				// No range loops, create new one.
				ctx.rl = NewRangeLoop(r, nodes, ctx)
				rl = ctx.rl
			} else {
				// Move forward over the list while new RL will found.
				crl := ctx.rl
				for {
					if crl.stat == rlFree {
						// Found it.
						rl = crl
						break
					}
					if crl.stat != rlFree {
						// RL is in use, need to go deeper.
						if crl.next != nil {
							crl = crl.next
							continue
						} else {
							// End of the list, create new free RL and exit from the loop.
							crl.next = NewRangeLoop(r, nodes, ctx)
							rl = crl.next
							break
						}
					}
				}
				// Prepare RL object.
				rl.cntr = 0
				rl.n = r
				rl.nodes = nodes
				rl.ctx = ctx
			}
			// Mark RL as inuse and loop over var using inspector.
			rl.stat = rlInuse
			ctx.Err = v.ins.Loop(v.val, rl, &ctx.buf, ctx.bufS[1:]...)
			rl.stat = rlFree
			return
		}
	}
}

func (ctx *Ctx) replaceQB(path []byte) []byte {
	qbLi := bytes.Index(path, qbL)
	qbRi := bytes.Index(path, qbR)
	if qbLi != -1 && qbRi != -1 && qbLi < qbRi && qbRi < len(path) {
		ctx.BufAcc.StakeOut()
		ctx.BufAcc.Write(path[0:qbLi]).Write(dot)
		ctx.chQB = false
		ctx.bufX = ctx.get(path[qbLi+1:qbRi], nil)
		if ctx.bufX != nil {
			if err := ctx.BufAcc.WriteX(ctx.bufX).Error(); err != nil {
				ctx.Err = err
				ctx.chQB = true
				return nil
			}
		}
		ctx.chQB = true
		ctx.BufAcc.Write(path[qbRi+1:])
		path = ctx.BufAcc.StakedBytes()
	}
	return path
}

// Compare method.
func (ctx *Ctx) cmp(path []byte, cond op, right []byte) bool {
	// Split path.
	ctx.bufS = ctx.bufS[:0]
	ctx.bufS = bytealg.AppendSplitString(ctx.bufS, byteconv.B2S(path), ".", -1)
	if len(ctx.bufS) == 0 {
		return false
	}

	for i := 0; i < ctx.ln; i++ {
		v := &ctx.vars[i]
		if v.key == ctx.bufS[0] {
			// Compare var with right value using inspector.
			ctx.Err = v.ins.Compare(v.val, inspector.Op(cond), byteconv.B2S(right), &ctx.bufBl, ctx.bufS[1:]...)
			if ctx.Err != nil {
				return false
			}
			return ctx.bufBl
		}
	}

	return false
}

// Reset the context.
//
// Made to use together with pools.
func (ctx *Ctx) Reset() {
	for i := 0; i < ctx.ln; i++ {
		ctx.vars[i].val = nil
	}
	ctx.ln = 0

	for i := 0; i < ctx.lenBB; i++ {
		ctx.bufBB[i] = ctx.bufBB[i][:0]
	}
	ctx.lenBB = 0

	for i := 0; i < ctx.ipvl; i++ {
		_ = ipoolRegistry.release(ctx.ipv[i].key, ctx.ipv[i].val)
		ctx.ipv[i].key, ctx.ipv[i].val = "", nil
	}
	ctx.ipvl = 0

	ctx.Err = nil
	ctx.bufX = nil
	ctx.accB = ctx.accB[:0]
	ctx.buf = ctx.buf[:0]
	ctx.bufS = ctx.bufS[:0]
	ctx.bufA = ctx.bufA[:0]
	ctx.bufLC = ctx.bufLC[:0]
	ctx.bufI, ctx.bufI_ = 0, 0
	ctx.BufAcc.Reset()
	ctx.Buf.Reset()
	ctx.Buf1.Reset()
	ctx.Buf2.Reset()

	ctx.brkD = 0
	ctx.rl.Reset()
}
