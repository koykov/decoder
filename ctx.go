package decoder

import (
	"github.com/koykov/bytealg"
	"github.com/koykov/fastconv"
	"github.com/koykov/inspector"
	"github.com/koykov/vector"
)

// Context object. Contains list of variables that can be used as source or destination.
type Ctx struct {
	// List of context variables and list len.
	vars []ctxVar
	ln   int
	// Vector parsers list and list len.
	p  []vector.Interface
	pl int
	// Internal buffers.
	accB []byte
	buf  []byte
	bufS []string
	bufI int
	bufX interface{}
	bufA []interface{}

	// External buffers to use in modifier and other callbacks.
	Buf, Buf1, Buf2 bytealg.ChainBuf

	Err error
}

// Context variable object.
type ctxVar struct {
	key  string
	val  interface{}
	ins  inspector.Inspector
	node *vector.Node
}

// Make new context object.
func NewCtx() *Ctx {
	ctx := Ctx{
		vars: make([]ctxVar, 0),
		bufS: make([]string, 0),
		bufA: make([]interface{}, 0),
	}
	return &ctx
}

// Set the variable to context.
// Inspector ins should be correspond to variable val.
func (c *Ctx) Set(key string, val interface{}, ins inspector.Inspector) {
	for i := 0; i < c.ln; i++ {
		if c.vars[i].key == key {
			// Update existing variable.
			c.vars[i].val = val
			c.vars[i].ins = ins
			return
		}
	}
	// Add new variable.
	if c.ln < len(c.vars) {
		// Use existing item in variable list..
		c.vars[c.ln].key = key
		c.vars[c.ln].val = val
		c.vars[c.ln].ins = ins
		c.vars[c.ln].node = nil
	} else {
		// Extend the variable list with new one.
		c.vars = append(c.vars, ctxVar{
			key: key,
			val: val,
			ins: ins,
		})
	}
	// Increase variables count.
	c.ln++
}

// Set static variable to context.
func (c *Ctx) SetStatic(key string, val interface{}) {
	ins, err := inspector.GetInspector("static")
	if err != nil {
		c.Err = err
		return
	}
	c.Set(key, val, ins)
}

// Parse source data and set it to context as key.
func (c *Ctx) SetVector(key string, data []byte, typ VectorType) (vec vector.Interface, err error) {
	vec = c.getParser(typ)
	if err = vec.Parse(data); err != nil {
		return
	}
	node := vec.Root()
	err = c.SetVectorNode(key, node)
	return
}

// Directly set node to context as key.
func (c *Ctx) SetVectorNode(key string, node *vector.Node) error {
	if node == nil || node.Type() == vector.TypeNull {
		return ErrEmptyNode
	}
	for i := 0; i < c.ln; i++ {
		if c.vars[i].key == key {
			// Update existing variable.
			c.vars[i].node = node
			c.vars[i].val, c.vars[i].ins = nil, nil
			return nil
		}
	}
	// Add new variable.
	if c.ln < len(c.vars) {
		// Use existing item in variable list..
		c.vars[c.ln].key = key
		c.vars[c.ln].node = node
		c.vars[c.ln].val, c.vars[c.ln].ins = nil, nil
	} else {
		// Extend the variable list with new one.
		c.vars = append(c.vars, ctxVar{
			key:  key,
			node: node,
		})
	}
	// Increase variables count.
	c.ln++
	return nil
}

// Get arbitrary value from the context by path.
//
// See Ctx.get().
// Path syntax: <ctxVrName>[.<Field>[.<NestedField0>[....<NestedFieldN>]]]
// Examples:
// * user.Bio.Birthday
// * staticVar
func (c *Ctx) Get(path string) interface{} {
	return c.get(fastconv.S2B(path), nil)
}

// Return accumulative buffer.
func (c *Ctx) AcquireBytes() []byte {
	return c.accB
}

// Update accumulative buffer with p.
func (c *Ctx) ReleaseBytes(p []byte) {
	if len(p) == 0 {
		return
	}
	c.accB = p
}

// Internal getter.
func (c *Ctx) get(path []byte, subset [][]byte) interface{} {
	if len(path) == 0 {
		return nil
	}

	// Split path to separate words using dot as separator.
	// So, path user.Bio.Birthday will convert to []string{"user", "Bio", "Birthday"}
	c.bufS = c.bufS[:0]
	c.bufS = bytealg.AppendSplitStr(c.bufS, fastconv.B2S(path), ".", -1)
	if len(c.bufS) == 0 {
		return nil
	}

	// Look for first path chunk in vars.
	for i, v := range c.vars {
		if i == c.ln {
			// Vars limit reached, exit.
			break
		}
		if v.key == c.bufS[0] {
			// Var found.
			if v.node != nil {
				// Var is JSON node.
				if len(subset) > 0 {
					// List of subsets provided.
					// Preserve item in []str buffer to check each key separately.
					c.bufS = append(c.bufS, "")
					for _, tail := range subset {
						if len(tail) > 0 {
							// Fill preserved item with subset's value.
							c.bufS[len(c.bufS)-1] = fastconv.B2S(tail)
							c.bufX = v.node.Get(c.bufS[1:]...)
							if n, ok := c.bufX.(*vector.Node); ok && n.Type() != vector.TypeNull {
								// Successful hunt.
								break
							}
						}
					}
				} else {
					c.bufX = v.node.Get(c.bufS[1:]...)
				}
				return c.bufX
			}
			if v.ins != nil {
				// Variable is covered by inspector.
				c.Err = v.ins.GetTo(v.val, &c.bufX, c.bufS[1:]...)
				if c.Err != nil {
					return nil
				}
				return c.bufX
			}
			return v.val
		}
	}
	return nil
}

// Internal setter.
//
// Set val to destination by address path.
func (c *Ctx) set(path []byte, val interface{}, insName []byte) error {
	if len(path) == 0 {
		return nil
	}
	c.bufS = c.bufS[:0]
	c.bufS = bytealg.AppendSplitStr(c.bufS, fastconv.B2S(path), ".", -1)
	if len(c.bufS) == 0 {
		return nil
	}
	if c.bufS[0] == "ctx" || c.bufS[0] == "context" {
		if len(c.bufS) == 1 {
			// Attempt to overwrite the whole context object caught.
			return nil
		}
		// Var-to-ctx case.
		ctxPath := fastconv.B2S(path[len(c.bufS[0])+1:])
		if len(insName) > 0 {
			ins, err := inspector.GetInspector(fastconv.B2S(insName))
			if err != nil {
				return err
			}
			c.Set(ctxPath, val, ins)
		} else if node, ok := val.(*vector.Node); ok {
			_ = c.SetVectorNode(ctxPath, node)
		} else {
			c.SetStatic(ctxPath, val)
		}
		return nil
	}
	// Var-to-var case.
	for i, v := range c.vars {
		if i == c.ln {
			break
		}
		if v.key == c.bufS[0] {
			if v.ins != nil {
				c.bufX = val
				c.Err = v.ins.SetWB(v.val, c.bufX, c, c.bufS[1:]...)
				if c.Err != nil {
					return c.Err
				}
			}
			break
		}
	}
	return nil
}

// Get new JSON vector object from the context buffer.
func (c *Ctx) getParser(typ VectorType) (v vector.Interface) {
	if c.pl < len(c.p) {
		v = ensureHelper(c.p[c.pl], typ)
	} else {
		v = newVector(typ)
		c.p = append(c.p, v)
	}
	c.pl++
	return v
}

// Reset the context.
//
// Made to use together with pools.
func (c *Ctx) Reset() {
	for i := 0; i < c.ln; i++ {
		c.vars[i].node = nil
	}
	c.ln = 0

	for i := 0; i < c.pl; i++ {
		c.p[i].Reset()
	}
	c.pl = 0

	c.Err = nil
	c.bufX = nil
	c.accB = c.accB[:0]
	c.buf = c.buf[:0]
	c.bufS = c.bufS[:0]
	c.bufA = c.bufA[:0]
	c.Buf.Reset()
	c.Buf1.Reset()
	c.Buf2.Reset()
}
