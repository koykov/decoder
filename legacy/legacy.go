package legacy

import (
	"github.com/koykov/decoder"
	_ "github.com/koykov/decoder_vector" // require to add support of vector pools
	"github.com/koykov/fastconv"
	"github.com/koykov/vector"
)

// Parse JSON source and register it in the ctx.
func cbJsonParse(ctx *decoder.Ctx, args []any) (err error) {
	return cbParse(ctx, args, "jsonvector")
}

// Parse URL source and register it in the ctx.
func cbUrlParse(ctx *decoder.Ctx, args []any) (err error) {
	return cbParse(ctx, args, "urlvector")
}

// Parse XML source and register it in the ctx.
func cbXmlParse(ctx *decoder.Ctx, args []any) (err error) {
	return cbParse(ctx, args, "xmlvector")
}

// Parse YAML source and register it in the ctx.
func cbYamlParse(ctx *decoder.Ctx, args []any) (err error) {
	return cbParse(ctx, args, "yamlvector")
}

// Parse HAL source and register it in the ctx.
func cbHalParse(ctx *decoder.Ctx, args []any) (err error) {
	return cbParse(ctx, args, "halvector")
}

// Parse source of type and register it in the ctx.
//
// Takes two arguments, the first must contain JSON text, the second - key to register parsed json in the ctx.
// Example of usage:
// <code>jsonParse("{\"a\":\"foo\"}", "parsed0")
// or
// <code>jsonParse(jsonSrc, "parsed1")</code>
// , where jsonSrc contains "{\"b\":[true,true,false]}".
func cbParse(ctx *decoder.Ctx, args []any, ipool string) (err error) {
	if len(args) < 2 {
		return decoder.ErrCbPoorArgs
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
			var vraw any
			if vraw, err = ctx.AcquireFrom(ipool); err != nil {
				return
			}
			var vec vector.Interface
			if vec, ok = vraw.(vector.Interface); !ok || vec == nil {
				return
			}
			if err = vec.Parse(src); err != nil {
				return err
			}
			err = ctx.SetVectorNode(fastconv.B2S(*key), vec.Root())
		}
	}
	return
}
