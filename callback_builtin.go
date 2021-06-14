package decoder

import (
	"github.com/koykov/fastconv"
	"github.com/koykov/vector"
)

// Example of callback function to use in parser tests.
func cbFoo(_ *Ctx, _ []interface{}) error {
	return nil
}

// Parse json source and register it in the ctx.
func cbJsonParse(ctx *Ctx, args []interface{}) (err error) {
	return cbParse(ctx, args, VectorJson)
}

// Parse json source and register it in the ctx.
func cbUrlParse(ctx *Ctx, args []interface{}) (err error) {
	return cbParse(ctx, args, VectorUrl)
}

// Parse json source and register it in the ctx.
func cbXmlParse(ctx *Ctx, args []interface{}) (err error) {
	return cbParse(ctx, args, VectorXml)
}

// Parse json source and register it in the ctx.
func cbYamlParse(ctx *Ctx, args []interface{}) (err error) {
	return cbParse(ctx, args, VectorYaml)
}

// Parse source of type and register it in the ctx.
//
// Takes two arguments, the first should contains JSON text, the second - key to register parsed json in the ctx.
// Example of usage:
// <code>jsonParse("{\"a\":\"foo\"}", "parsed0")
// or
// <code>jsonParse(jsonSrc, "parsed1")</code>
// , where jsonSrc contains "{\"b\":[true,true,false]}".
func cbParse(ctx *Ctx, args []interface{}, typ VectorType) (err error) {
	if len(args) < 2 {
		return ErrCbPoorArgs
	}
	var src []byte
	switch args[0].(type) {
	case *[]byte:
		src = *args[0].(*[]byte)
	case []byte:
		src = args[0].([]byte)
	case *string:
		src = fastconv.S2B(*args[0].(*string))
	case string:
		src = fastconv.S2B(args[0].(string))
	case *vector.Node:
		node := args[0].(*vector.Node)
		if node.Type() == vector.TypeStr {
			src = node.Bytes()
		}
	}
	if len(src) > 0 {
		if key, ok := args[1].(*[]byte); ok {
			_, err = ctx.SetVector(fastconv.B2S(*key), src, typ)
		}
	}
	return
}
