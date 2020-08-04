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

	p *jsonvector.Vector

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
		p:    jsonvector.NewVector(),
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
	if err = c.p.Parse(data); err != nil {
		return
	}
	jsn := c.p.Get()

	for i := 0; i < c.ln; i++ {
		if c.vars[i].key == key {
			// Update existing variable.
			c.vars[i].jsn = jsn
			c.vars[i].val, c.vars[i].ins = nil, nil
			return
		}
	}
	// Add new variable.
	if c.ln < len(c.vars) {
		// Use existing item in variable list..
		c.vars[c.ln].key = key
		c.vars[c.ln].jsn = jsn
		c.vars[c.ln].val, c.vars[c.ln].ins = nil, nil
	} else {
		// Extend the variable list with new one.
		c.vars = append(c.vars, ctxVar{
			key: key,
			jsn: jsn,
		})
	}
	// Increase variables count.
	c.ln++
	return
}

func (c *Ctx) get(path []byte) interface{} {
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
				c.bufX = v.jsn.Get(c.bufS[1:]...)
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
		}
	}
	return nil
}

func (c *Ctx) Reset() {
	c.p.Reset()
	for i := 0; i < c.ln; i++ {
		c.vars[i].jsn = nil
	}
	c.ln = 0
	c.Err = nil
	c.bufX = nil
	c.buf = c.buf[:0]
	c.bufS = c.bufS[:0]
	c.bufA = c.bufA[:0]
}
