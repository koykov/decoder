package jsondecoder

import (
	"github.com/koykov/fastconv"
	"github.com/koykov/jsonvector"
)

func cbFoo(_ *Ctx, _ []interface{}) error {
	return nil
}

func cbJsonParse(ctx *Ctx, args []interface{}) (err error) {
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
	case *jsonvector.Node:
		node := args[0].(*jsonvector.Node)
		if node.Type() == jsonvector.TypeStr {
			src = node.Bytes()
		}
	}
	if len(src) > 0 {
		if key, ok := args[1].(*[]byte); ok {
			_, err = ctx.SetJson(fastconv.B2S(*key), src)
		}
	}
	return
}
