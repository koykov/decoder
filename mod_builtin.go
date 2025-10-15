package decoder

import (
	"bytes"

	"github.com/koykov/vector"
)

// Replace empty val with default value.
func modDefault(ctx *Ctx, buf *any, val any, args []any) (err error) {
	// Check val is empty.
	var empty_ bool
	switch x := val.(type) {
	case *[]byte:
		empty_ = len(*x) == 0
	case []byte:
		empty_ = len(x) == 0
	case *string:
		empty_ = len(*x) == 0
	case string:
		empty_ = len(x) == 0
	case *bool:
		empty_ = !(*x)
	case bool:
		empty_ = !x
	case int:
		empty_ = x == 0
	case *int:
		empty_ = *x == 0
	case int8:
		empty_ = x == 0
	case *int8:
		empty_ = *x == 0
	case int16:
		empty_ = x == 0
	case *int16:
		empty_ = *x == 0
	case int32:
		empty_ = x == 0
	case *int32:
		empty_ = *x == 0
	case int64:
		empty_ = x == 0
	case *int64:
		empty_ = *x == 0
	case uint:
		empty_ = x == 0
	case *uint:
		empty_ = *x == 0
	case uint8:
		empty_ = x == 0
	case *uint8:
		empty_ = *x == 0
	case uint16:
		empty_ = x == 0
	case *uint16:
		empty_ = *x == 0
	case uint32:
		empty_ = x == 0
	case *uint32:
		empty_ = *x == 0
	case uint64:
		empty_ = x == 0
	case *uint64:
		empty_ = *x == 0
	case float32:
		empty_ = x == 0
	case *float32:
		empty_ = *x == 0
	case float64:
		empty_ = x == 0
	case *float64:
		empty_ = *x == 0
	case *vector.Node:
		empty_ = x.Type() == vector.TypeNull || x.Limit() == 0
	default:
		empty_ = false
	}
	if !empty_ {
		// Non-empty case - exiting.
		return
	}
	if len(args) == 0 {
		err = ErrModPoorArgs
	}

	// Implement default mod logic.
	switch x := args[0].(type) {
	case *[]byte:
		i := ctx.reserveBB()
		ctx.bufBB[i] = append(ctx.bufBB[i], *x...)
		*buf = &ctx.bufBB[i]
	case []byte:
		*buf = &x
	case *string:
		*buf = x
	case string:
		*buf = &x
	case *bool:
		*buf = x
	case bool:
		*buf = &x
	case int:
		*buf = &x
	case *int:
		*buf = x
	case int8:
		*buf = &x
	case *int8:
		*buf = x
	case int16:
		*buf = &x
	case *int16:
		*buf = x
	case int32:
		*buf = &x
	case *int32:
		*buf = x
	case int64:
		*buf = &x
	case *int64:
		*buf = x
	case uint:
		*buf = &x
	case *uint:
		*buf = x
	case uint8:
		*buf = &x
	case *uint8:
		*buf = x
	case uint16:
		*buf = &x
	case *uint16:
		*buf = x
	case uint32:
		*buf = &x
	case *uint32:
		*buf = x
	case uint64:
		*buf = &x
	case *uint64:
		*buf = x
	case float32:
		*buf = &x
	case *float32:
		*buf = x
	case float64:
		*buf = &x
	case *float64:
		*buf = x
	case *vector.Node:
		if x != nil {
			i := ctx.reserveBB()
			ctx.bufBB[i] = append(ctx.bufBB[i], x.Bytes()...)
			*buf = &ctx.bufBB[i]
		}
	default:
		*buf = nil
	}
	return
}

func modNewHidden(_ *Ctx, buf *any, val any, args []any) (err error) {
	// todo implement me
	return
}

// Conditional assignment modifier.
func modIfThen(_ *Ctx, buf *any, val any, args []any) (err error) {
	if len(args) == 0 {
		err = ErrModNoArgs
		return
	}
	if checkTrue(val) {
		*buf = args[0]
	}
	return
}

// Extended conditional assignment modifier (includes else case).
func modIfThenElse(_ *Ctx, buf *any, val any, args []any) (err error) {
	if len(args) < 2 {
		err = ErrModPoorArgs
		return
	}
	if checkTrue(val) {
		*buf = args[0]
	} else {
		*buf = args[1]
	}
	return
}

// Check if given val is a true.
func checkTrue(val any) (r bool) {
	switch x := val.(type) {
	case *[]byte:
		r = bytes.Equal(*x, bTrue)
	case []byte:
		r = bytes.Equal(x, bTrue)
	case *string:
		r = *x == "true"
	case string:
		r = x == "true"
	case *bool:
		r = *x
	case bool:
		r = x
	case int:
		r = x == 1
	case *int:
		r = *x == 1
	case int8:
		r = x == 1
	case *int8:
		r = *x == 1
	case int16:
		r = x == 1
	case *int16:
		r = *x == 1
	case int32:
		r = x == 1
	case *int32:
		r = *x == 1
	case int64:
		r = x == 1
	case *int64:
		r = *x == 1
	case uint:
		r = x == 1
	case *uint:
		r = *x == 1
	case uint8:
		r = x == 1
	case *uint8:
		r = *x == 1
	case uint16:
		r = x == 1
	case *uint16:
		r = *x == 1
	case uint32:
		r = x == 1
	case *uint32:
		r = *x == 1
	case uint64:
		r = x == 1
	case *uint64:
		r = *x == 1
	case float32:
		r = x == 1
	case *float32:
		r = *x == 1
	case float64:
		r = x == 1
	case *float64:
		r = *x == 1
	case *vector.Node:
		if x != nil {
			r = x.Bool()
		}
	default:
		r = false
	}
	return
}

var bTrue = []byte("true")
