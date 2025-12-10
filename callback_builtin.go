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
	var (
		raw  []byte
		path []string
	)
	switch x := args[0].(type) {
	case string:
		raw = byteconv.S2B(x)
	case *string:
		raw = byteconv.S2B(*x)
	case []byte:
		raw = x
	case *[]byte:
		raw = *x
	case []string:
		path = x
	case *[]string:
		path = *x
	default:
		return nil // cannot check path
	}

	if len(path) == 0 && len(raw) > 0 {
		ctx.splitPath(byteconv.B2S(raw), ".")
		if len(ctx.bufS) == 0 {
			return nil
		}
		path = ctx.bufS
	}
	if len(path) == 0 {
		return nil
	}

	var err error
	src, ins := ctx.get2(path[:1], nil)
	err = ins.Reset(src, path[1:]...)
	return err
}
