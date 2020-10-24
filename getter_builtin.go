package decoder

import (
	"hash/crc32"
	"strconv"
	"time"

	"github.com/koykov/fastconv"
	"github.com/koykov/inspector/testobj"
	"github.com/koykov/jsonvector"
)

// Calculate common crc32 hash of given arguments.
func getterCrc32(ctx *Ctx, buf *interface{}, args []interface{}) (err error) {
	if len(args) == 0 {
		err = ErrGetterPoorArgs
		return
	}
	ctx.Buf.Reset()
	for _, a := range args {
		switch a.(type) {
		case []byte:
			ctx.Buf.Write(a.([]byte))
		case *[]byte:
			ctx.Buf.Write(*a.(*[]byte))
		case string:
			ctx.Buf.WriteStr(a.(string))
		case *string:
			ctx.Buf.WriteStr(*a.(*string))
		case *jsonvector.Node:
			ctx.Buf.Write(a.(*jsonvector.Node).Bytes())
		}
	}
	if ctx.Buf.Len() > 0 {
		ctx.bufI = int(crc32.ChecksumIEEE(ctx.Buf.Bytes()))
		*buf = &ctx.bufI
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
		case *jsonvector.Node:
			hr.Cost, _ = args[1].(*jsonvector.Node).Float()
		}
		switch args[2].(type) {
		case *[]byte:
			hr.Comment = *args[2].(*[]byte)
		case *jsonvector.Node:
			hr.Comment = args[2].(*jsonvector.Node).Bytes()
		}
		*h = append(*h, hr)
		*buf = h
	}
	return
}
