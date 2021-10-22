package decoder

import (
	"hash/crc32"
	"strconv"
	"time"

	"github.com/koykov/fastconv"
	"github.com/koykov/inspector/testobj"
	"github.com/koykov/vector"
)

// Calculate common crc32 hash of given arguments.
func getterCrc32(ctx *Ctx, buf *interface{}, args []interface{}) (err error) {
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
			ctx.BufAcc.WriteStr(a.(string))
		case *string:
			ctx.BufAcc.WriteStr(*a.(*string))
		case *vector.Node:
			node := a.(*vector.Node)
			if node != nil {
				ctx.BufAcc.Write(node.Bytes())
			}
		}
	}
	if ctx.BufAcc.StakedLen() > 0 {
		ctx.bufI = int(crc32.ChecksumIEEE(ctx.BufAcc.StakedBytes()))
		*buf = &ctx.bufI
	}
	return
}

// Convert string to float.
func getterAtof(ctx *Ctx, buf *interface{}, args []interface{}) (err error) {
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
		raw = fastconv.B2S(args[0].([]byte))
	case []byte:
		raw = fastconv.B2S(*args[0].(*[]byte))
	default:
		ok = false
	}
	if ok {
		if ctx.bufF, err = strconv.ParseFloat(raw, 64); err == nil {
			*buf = &ctx.bufF
		}
	}
	return
}

// Example of getter callback that demonstrates how callbacks works.
func getterAppendTestHistory(_ *Ctx, buf *interface{}, args []interface{}) (err error) {
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
