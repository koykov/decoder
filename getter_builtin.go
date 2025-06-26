package decoder

import (
	"hash/crc32"
	"strconv"
	"time"

	"github.com/koykov/byteconv"
	"github.com/koykov/inspector/testobj"
	"github.com/koykov/vector"
	"github.com/koykov/x2bytes"
)

type atoxT int

const (
	atoi atoxT = iota
	atou
	atof
	atob
)

// Calculate common crc32 hash of given arguments.
func getterCrc32(ctx *Ctx, buf *any, args []any) (err error) {
	n := len(args)
	if n == 0 {
		err = ErrGetterPoorArgs
		return
	}
	ctx.BufAcc.StakeOut()
	_ = args[n-1]
	for i := 0; i < n; i++ {
		a := args[i]
		switch x := a.(type) {
		case []byte:
			ctx.BufAcc.Write(x)
		case *[]byte:
			ctx.BufAcc.Write(*x)
		case string:
			ctx.BufAcc.WriteString(x)
		case *string:
			ctx.BufAcc.WriteString(*x)
		case *vector.Node:
			if x != nil {
				ctx.BufAcc.Write(x.Bytes())
			}
		}
	}
	if ctx.BufAcc.StakedLen() > 0 {
		ctx.bufI = int64(crc32.ChecksumIEEE(ctx.BufAcc.StakedBytes()))
		*buf = &ctx.bufI
	}
	return
}

// Convert string to int.
func getterAtoi(ctx *Ctx, buf *any, args []any) (err error) {
	return atox(ctx, buf, args, atoi)
}

// Convert string to uint.
func getterAtou(ctx *Ctx, buf *any, args []any) (err error) {
	return atox(ctx, buf, args, atou)
}

// Convert string to float.
func getterAtof(ctx *Ctx, buf *any, args []any) (err error) {
	return atox(ctx, buf, args, atof)
}

// Convert string to bool.
func getterAtob(ctx *Ctx, buf *any, args []any) (err error) {
	return atox(ctx, buf, args, atob)
}

func atox(ctx *Ctx, buf *any, args []any, target atoxT) (err error) {
	if len(args) < 1 {
		err = ErrGetterPoorArgs
		return
	}
	var raw string
	ok := true
	switch x := args[0].(type) {
	case *vector.Node:
		raw = x.String()
	case string:
		raw = x
	case *string:
		raw = *x
	case *[]byte:
		raw = byteconv.B2S(*x)
	case []byte:
		raw = byteconv.B2S(x)
	default:
		ok = false
	}
	if ok {
		switch target {
		case atoi:
			if ctx.bufI, err = strconv.ParseInt(raw, 10, 64); err == nil {
				*buf = &ctx.bufI
			}
		case atou:
			if ctx.bufU, err = strconv.ParseUint(raw, 10, 64); err == nil {
				*buf = &ctx.bufU
			}
		case atof:
			if ctx.bufF, err = strconv.ParseFloat(raw, 64); err == nil {
				*buf = &ctx.bufF
			}
		case atob:
			if ctx.bufBl, err = strconv.ParseBool(raw); err == nil {
				*buf = &ctx.bufBl
			}
		}
	}
	return
}

func getterItoa(ctx *Ctx, buf *any, args []any) (err error) {
	if len(args) < 1 {
		err = ErrGetterPoorArgs
		return
	}
	i := ctx.reserveBB()
	if node, ok := args[0].(*vector.Node); ok {
		ctx.bufBB[i] = append(ctx.bufBB[i], node.ForceBytes()...)
		*buf = &ctx.bufBB[i]
	} else if ctx.bufBB[i], err = x2bytes.IntToBytes(ctx.bufBB[i], args[0]); err == nil {
		*buf = &ctx.bufBB[i]
	}
	return
}

func getterUtoa(ctx *Ctx, buf *any, args []any) (err error) {
	if len(args) < 1 {
		err = ErrGetterPoorArgs
		return
	}
	i := ctx.reserveBB()
	if node, ok := args[0].(*vector.Node); ok {
		ctx.bufBB[i] = append(ctx.bufBB[i], node.ForceBytes()...)
		*buf = &ctx.bufBB[i]
	} else if ctx.bufBB[i], err = x2bytes.UintToBytes(ctx.bufBB[i], args[0]); err == nil {
		*buf = &ctx.bufBB[i]
	}
	return
}

// Example of getter callback that demonstrates how callbacks works.
func getterAppendTestHistory(_ *Ctx, buf *any, args []any) (err error) {
	if len(args) < 3 {
		err = ErrGetterPoorArgs
		return
	}
	if h, ok := args[0].(*[]testobj.TestHistory); ok || args[0] == nil {
		if h == nil {
			h = &[]testobj.TestHistory{}
		}
		hr := testobj.TestHistory{DateUnix: time.Now().Unix()}
		switch x := args[1].(type) {
		case *[]byte:
			hr.Cost, _ = strconv.ParseFloat(byteconv.B2S(*x), 64)
		case *vector.Node:
			hr.Cost, _ = x.Float()
		}
		switch x := args[2].(type) {
		case *[]byte:
			hr.Comment = *x
		case *vector.Node:
			hr.Comment = x.Bytes()
		}
		*h = append(*h, hr)
		*buf = h
	}
	return
}
