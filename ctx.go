package jsondecoder

import (
	"github.com/koykov/inspector"
	"github.com/koykov/jsonvector"
)

type Ctx struct {
	p *jsonvector.Vector

	vars []ctxVar
	ln   int
}

type ctxVar struct {
	key string
	val interface{}
	ins inspector.Inspector
}
