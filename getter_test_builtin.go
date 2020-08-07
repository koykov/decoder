package jsondecoder

import (
	"strconv"
	"time"

	"github.com/koykov/fastconv"
	"github.com/koykov/inspector/testobj"
	"github.com/koykov/jsonvector"
)

func getterAppendTestHistory(ctx *Ctx, buf *interface{}, args []interface{}) (err error) {
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
			hr.Cost = args[1].(*jsonvector.Node).Float()
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
	return nil
}
