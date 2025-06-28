package decoder

import "github.com/koykov/byteconv"

func modFmtFormat(ctx *Ctx, buf *any, _ any, args []any) (err error) {
	if len(args) < 1 {
		err = ErrModPoorArgs
		return
	}
	var sfmt string
	switch x := args[0].(type) {
	case string:
		sfmt = x
	case *string:
		sfmt = *x
	case []byte:
		sfmt = byteconv.B2S(x)
	case *[]byte:
		sfmt = byteconv.B2S(*x)
	default:
		return nil
	}
	ctx.BufAcc.StakeOut().
		WriteFormat(sfmt, args[1:]...)
	i := ctx.reserveBB()
	ctx.bufBB[i] = ctx.BufAcc.StakedBytes()
	*buf = &ctx.bufBB[i]
	return
}
