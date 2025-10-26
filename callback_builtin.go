package decoder

import (
	"fmt"

	"github.com/koykov/byteconv"
)

func cbPrint(_ *Ctx, args []any) error {
	fmt.Print(args...)
	return nil
}

func cbPrintln(_ *Ctx, args []any) error {
	fmt.Println(args...)
	return nil
}

func cbReset(ctx *Ctx, args []any) error {
	if len(args) == 0 {
		return ErrModNoArgs
	}
	var path []byte
	switch x := args[0].(type) {
	case string:
		path = byteconv.S2B(x)
	case *string:
		path = byteconv.S2B(*x)
	case []byte:
		path = x
	case *[]byte:
		path = *x
	default:
		return nil // cannot check path
	}

	ctx.splitPath(byteconv.B2S(path), ".")
	if len(ctx.bufS) == 0 {
		return nil
	}

	var err error
	src, ins := ctx.get2(ctx.bufS[:1], nil)
	err = ins.Reset(src, ctx.bufS[1:]...)
	return err
}
