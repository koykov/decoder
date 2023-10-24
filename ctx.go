package decoder

import (
	"strings"

	"github.com/koykov/bytealg"
	"github.com/koykov/bytebuf"
	"github.com/koykov/fastconv"
	"github.com/koykov/inspector"
	"github.com/koykov/vector"
)

// Ctx represents decoder context object.
//
// Contains list of variables that can be used as source or destination.
type Ctx struct {
	// List of context variables and list len.
	vars []ctxVar
	ln   int
	// Vector parsers list and list len.
	p  [VectorsSupported][]vector.Interface
	pl [VectorsSupported]int
	// Internal buffers.
	accB  []byte
	buf   []byte
	bufBB [][]byte
	lenBB int
	bufS  []string
	bufI  int64
	bufU  uint64
	bufF  float64
	bufBl bool
	bufX  any
	bufA  []any

	// List of variables taken from ipools and registered to return back.
	ipv  []ipoolVar
	ipvl int

	// External buffers to use in modifier and condition helpers.
	BufAcc bytebuf.AccumulativeBuf
	// todo remove as unused later
	Buf, Buf1, Buf2 bytebuf.ChainBuf

	Err error
}

// Context variable object.
type ctxVar struct {
	key  string
	val  any
	ins  inspector.Inspector
	node *vector.Node
}

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
		// Use existing item in variable list..
		ctx.vars[ctx.ln].key = key
		ctx.vars[ctx.ln].val = val
		ctx.vars[ctx.ln].ins = ins
		ctx.vars[ctx.ln].node = nil
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

// SetVector parses source data and register it in context under given key.
func (ctx *Ctx) SetVector(key string, data []byte, typ VectorType) (vec vector.Interface, err error) {
	vec = ctx.getParser(typ)
	if err = vec.Parse(data); err != nil {
		return
	}
	node := vec.Root()
	err = ctx.SetVectorNode(key, node)
	return
}

// SetVectorNode directly registers node in context under given key.
func (ctx *Ctx) SetVectorNode(key string, node *vector.Node) error {
	if node == nil || node.Type() == vector.TypeNull {
		return ErrEmptyNode
	}
	for i := 0; i < ctx.ln; i++ {
		if ctx.vars[i].key == key {
			// Update existing variable.
			ctx.vars[i].node = node
			ctx.vars[i].val, ctx.vars[i].ins = nil, nil
			return nil
		}
	}
	// Add new variable.
	if ctx.ln < len(ctx.vars) {
		// Use existing item in variable list..
		ctx.vars[ctx.ln].key = key
		ctx.vars[ctx.ln].node = node
		ctx.vars[ctx.ln].val, ctx.vars[ctx.ln].ins = nil, nil
	} else {
		// Extend the variable list with new one.
		ctx.vars = append(ctx.vars, ctxVar{
			key:  key,
			node: node,
		})
	}
	// Increase variables count.
	ctx.ln++
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
	return ctx.get(fastconv.S2B(path), nil)
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
	return fastconv.B2S(ctx.accB[off:])
}

// AcquireFrom receives new variable from given pool and register it to return batch after finish template processing.
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
	if len(path) == 0 {
		return nil
	}

	// Split path to separate words using dot as separator.
	ctx.splitPath(fastconv.B2S(path), ".")
	if len(ctx.bufS) == 0 {
		return nil
	}

	// Look for first path chunk in vars.
	for i, v := range ctx.vars {
		if i == ctx.ln {
			// Vars limit reached, exit.
			break
		}
		if v.key == ctx.bufS[0] {
			// Var found.
			if v.node != nil {
				// Var is JSON node.
				if len(subset) > 0 {
					// List of subsets provided.
					// Preserve item in []str buffer to check each key separately.
					ctx.bufS = append(ctx.bufS, "")
					for _, tail := range subset {
						if len(tail) > 0 {
							// Fill preserved item with subset's value.
							ctx.bufS[len(ctx.bufS)-1] = fastconv.B2S(tail)
							ctx.bufX = v.node.Get(ctx.bufS[1:]...)
							if n, ok := ctx.bufX.(*vector.Node); ok && n.Type() != vector.TypeNull {
								// Successful hunt.
								break
							}
						}
					}
				} else {
					ctx.bufX = v.node.Get(ctx.bufS[1:]...)
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
	if len(path) == 0 {
		return nil
	}
	ctx.bufS = ctx.bufS[:0]
	ctx.bufS = bytealg.AppendSplit(ctx.bufS, fastconv.B2S(path), ".", -1)
	if len(ctx.bufS) == 0 {
		return nil
	}
	if ctx.bufS[0] == "ctx" || ctx.bufS[0] == "context" {
		if len(ctx.bufS) == 1 {
			// Attempt to overwrite the whole context object caught.
			return nil
		}
		// Var-to-ctx case.
		ctxPath := fastconv.B2S(path[len(ctx.bufS[0])+1:])
		if len(insName) > 0 {
			ins, err := inspector.GetInspector(fastconv.B2S(insName))
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
	for i, v := range ctx.vars {
		if i == ctx.ln {
			break
		}
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

// Get new JSON vector object from the context buffer.
func (ctx *Ctx) getParser(typ VectorType) (v vector.Interface) {
	if ctx.pl[typ] < len(ctx.p[typ]) {
		v = ensureHelper(ctx.p[typ][ctx.pl[typ]], typ)
	} else {
		v = newVector(typ)
		ctx.p[typ] = append(ctx.p[typ], v)
	}
	ctx.pl[typ]++
	return v
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

// Reset the context.
//
// Made to use together with pools.
func (ctx *Ctx) Reset() {
	for i := 0; i < ctx.ln; i++ {
		ctx.vars[i].node = nil
	}
	ctx.ln = 0

	for i := 0; i < VectorsSupported; i++ {
		for j := 0; j < ctx.pl[i]; j++ {
			ctx.p[i][j].Reset()
		}
		ctx.pl[i] = 0
	}

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
	ctx.BufAcc.Reset()
	ctx.Buf.Reset()
	ctx.Buf1.Reset()
	ctx.Buf2.Reset()
}
