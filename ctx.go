package jsondecoder

import (
	"github.com/koykov/bytealg"
	"github.com/koykov/fastconv"
	"github.com/koykov/inspector"
	"github.com/koykov/jsonvector"
)

type Ctx struct {
	vars []ctxVar
	ln   int

	p  []*jsonvector.Vector
	pl int

	buf  []byte
	bufX interface{}
	bufS []string
	bufA []interface{}

	Err error
}

type ctxVar struct {
	key string
	val interface{}
	ins inspector.Inspector
	jsn *jsonvector.Node
}

func NewCtx() *Ctx {
	ctx := Ctx{
		vars: make([]ctxVar, 0),
		bufS: make([]string, 0),
		bufA: make([]interface{}, 0),
	}
	return &ctx
}

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
		c.vars[c.ln].jsn = nil
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

func (c *Ctx) SetJson(key string, data []byte) (err error) {
	vec := c.getParser()
	if err = vec.Parse(data); err != nil {
		return
	}
	node := vec.Get()
	err = c.SetJsonNode(key, node)
	return
}

func (c *Ctx) SetJsonNode(key string, node *jsonvector.Node) error {
	if node == nil {
		return ErrEmptyNode
	}
	for i := 0; i < c.ln; i++ {
		if c.vars[i].key == key {
			// Update existing variable.
			c.vars[i].jsn = node
			c.vars[i].val, c.vars[i].ins = nil, nil
			return nil
		}
	}
	// Add new variable.
	if c.ln < len(c.vars) {
		// Use existing item in variable list..
		c.vars[c.ln].key = key
		c.vars[c.ln].jsn = node
		c.vars[c.ln].val, c.vars[c.ln].ins = nil, nil
	} else {
		// Extend the variable list with new one.
		c.vars = append(c.vars, ctxVar{
			key: key,
			jsn: node,
		})
	}
	// Increase variables count.
	c.ln++
	return nil
}

func (c *Ctx) get(path []byte, subset [][]byte) interface{} {
	if len(path) == 0 {
		return nil
	}
	c.bufS = c.bufS[:0]
	c.bufS = bytealg.AppendSplitStr(c.bufS, fastconv.B2S(path), ".", -1)
	for i, v := range c.vars {
		if i == c.ln {
			break
		}
		if v.key == c.bufS[0] {
			if v.jsn != nil {
				if len(subset) > 0 {
					c.bufS = append(c.bufS, "")
					for _, tail := range subset {
						if len(tail) > 0 {
							c.bufS[len(c.bufS)-1] = fastconv.B2S(tail)
							if c.bufX = v.jsn.Get(c.bufS[1:]...); c.bufX != nil {
								break
							}
						}
					}
				} else {
					c.bufX = v.jsn.Get(c.bufS[1:]...)
				}
				return c.bufX
			}
			if v.ins != nil {
				c.Err = v.ins.GetTo(v.val, &c.bufX, c.bufS[1:]...)
				if c.Err != nil {
					return nil
				}
				return c.bufX
			}
		}
	}
	return nil
}

func (c *Ctx) set(path []byte, val interface{}) error {
	if len(path) == 0 {
		return nil
	}
	c.bufS = c.bufS[:0]
	c.bufS = bytealg.AppendSplitStr(c.bufS, fastconv.B2S(path), ".", -1)
	for i, v := range c.vars {
		if i == c.ln {
			break
		}
		if v.key == c.bufS[0] {
			if v.ins != nil {
				c.bufX = val
				c.Err = v.ins.Set(v.val, c.bufX, c.bufS[1:]...)
				if c.Err != nil {
					return c.Err
				}
			}
			break
		}
	}
	return nil
}

func (c *Ctx) getParser() *jsonvector.Vector {
	var v *jsonvector.Vector
	if c.pl < len(c.p) {
		v = c.p[c.pl]
	} else {
		v = jsonvector.NewVector()
		c.p = append(c.p, v)
	}
	c.pl++
	return v
}

func (c *Ctx) Reset() {
	for i := 0; i < c.ln; i++ {
		c.vars[i].jsn = nil
	}
	c.ln = 0

	for i := 0; i < c.pl; i++ {
		c.p[i].Reset()
	}
	c.pl = 0

	c.Err = nil
	c.bufX = nil
	c.buf = c.buf[:0]
	c.bufS = c.bufS[:0]
	c.bufA = c.bufA[:0]
}
