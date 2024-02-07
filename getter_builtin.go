package decoder

import (
	"hash/crc32"
	"strconv"
	"time"

	"github.com/koykov/fastconv"
	"github.com/koykov/inspector/testobj"
	"github.com/koykov/vector"
	"github.com/koykov/x2bytes"
)

type target int

const (
	atoi target = iota
	atou
	atof
	atob
)

// Calculate common crc32 hash of given arguments.
func getterCrc32(ctx *Ctx, buf *any, args []any) (err error) {
	if len(args) == 0 {
		err = ErrGetterPoorArgs
		return
	}
	ctx.BufAcc.StakeOut()
	for _, a := range args {
		switch a.(type) {
		case []byte:
			ctx.BufAcc.Write(a.([]byte))
		case *[]byte:
			ctx.BufAcc.Write(*a.(*[]byte))
		case string:
			ctx.BufAcc.WriteString(a.(string))
		case *string:
			ctx.BufAcc.WriteString(*a.(*string))
		case *vector.Node:
			node := a.(*vector.Node)
			if node != nil {
				ctx.BufAcc.Write(node.Bytes())
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

func atox(ctx *Ctx, buf *any, args []any, target target) (err error) {
	if len(args) < 1 {
		err = ErrGetterPoorArgs
		return
	}
	var raw string
	ok := true
	switch args[0].(type) {
	case *vector.Node:
		raw = args[0].(*vector.Node).String()
	case string:
		raw = args[0].(string)
	case *string:
		raw = *args[0].(*string)
	case *[]byte:
		raw = fastconv.B2S(*args[0].(*[]byte))
	case []byte:
		raw = fastconv.B2S(args[0].([]byte))
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
		switch args[1].(type) {
		case *[]byte:
			hr.Cost, _ = strconv.ParseFloat(fastconv.B2S(*args[1].(*[]byte)), 64)
		case *vector.Node:
			hr.Cost, _ = args[1].(*vector.Node).Float()
		}
		switch args[2].(type) {
		case *[]byte:
			hr.Comment = *args[2].(*[]byte)
		case *vector.Node:
			hr.Comment = args[2].(*vector.Node).Bytes()
		}
		*h = append(*h, hr)
		*buf = h
	}
	return
}
