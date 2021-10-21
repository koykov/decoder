package decoder

import (
	"bytes"

	"github.com/koykov/vector"
)

var (
	bTrue = []byte("true")
)

// Replace empty val with default value.
func modDefault(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	// Check val is empty.
	var empty bool
	switch val.(type) {
	case *[]byte:
		a := *val.(*[]byte)
		empty = len(a) == 0
	case []byte:
		a := val.([]byte)
		empty = len(a) == 0
	case *string:
		a := *val.(*string)
		empty = len(a) == 0
	case string:
		a := val.(string)
		empty = len(a) == 0
	case *bool:
		a := *val.(*bool)
		empty = !a
	case bool:
		a := val.(bool)
		empty = !a
	case int:
		a := val.(int)
		empty = a == 0
	case *int:
		a := *val.(*int)
		empty = a == 0
	case int8:
		a := val.(int8)
		empty = a == 0
	case *int8:
		a := *val.(*int8)
		empty = a == 0
	case int16:
		a := val.(int16)
		empty = a == 0
	case *int16:
		a := *val.(*int16)
		empty = a == 0
	case int32:
		a := val.(int32)
		empty = a == 0
	case *int32:
		a := *val.(*int32)
		empty = a == 0
	case int64:
		a := val.(int64)
		empty = a == 0
	case *int64:
		a := *val.(*int64)
		empty = a == 0
	case uint:
		a := val.(uint)
		empty = a == 0
	case *uint:
		a := *val.(*uint)
		empty = a == 0
	case uint8:
		a := val.(uint8)
		empty = a == 0
	case *uint8:
		a := *val.(*uint8)
		empty = a == 0
	case uint16:
		a := val.(uint16)
		empty = a == 0
	case *uint16:
		a := *val.(*uint16)
		empty = a == 0
	case uint32:
		a := val.(uint32)
		empty = a == 0
	case *uint32:
		a := *val.(*uint32)
		empty = a == 0
	case uint64:
		a := val.(uint64)
		empty = a == 0
	case *uint64:
		a := *val.(*uint64)
		empty = a == 0
	case float32:
		a := val.(float32)
		empty = a == 0
	case *float32:
		a := *val.(*float32)
		empty = a == 0
	case float64:
		a := val.(float64)
		empty = a == 0
	case *float64:
		a := *val.(*float64)
		empty = a == 0
	case *vector.Node:
		node := val.(*vector.Node)
		empty = node.Type() == vector.TypeNull || node.Limit() == 0
	default:
		empty = false
	}
	if !empty {
		// Non-empty case - exiting.
		return
	}
	if len(args) == 0 {
		err = ErrModPoorArgs
	}

	// Implement default mod logic.
	switch args[0].(type) {
	case *[]byte:
		i := ctx.reserveBB()
		ctx.bufBB[i] = append(ctx.bufBB[i], *args[0].(*[]byte)...)
		*buf = &ctx.bufBB[i]
	case []byte:
		a := args[0].([]byte)
		*buf = &a
	case *string:
		a := *args[0].(*string)
		*buf = &a
	case string:
		a := args[0].(string)
		*buf = &a
	case *bool:
		a := *args[0].(*bool)
		*buf = &a
	case bool:
		a := args[0].(bool)
		*buf = &a
	case int:
		a := args[0].(int)
		*buf = &a
	case *int:
		a := *args[0].(*int)
		*buf = &a
	case int8:
		a := args[0].(int8)
		*buf = &a
	case *int8:
		a := *args[0].(*int8)
		*buf = &a
	case int16:
		a := args[0].(int16)
		*buf = &a
	case *int16:
		a := *args[0].(*int16)
		*buf = &a
	case int32:
		a := args[0].(int32)
		*buf = &a
	case *int32:
		a := *args[0].(*int32)
		*buf = &a
	case int64:
		a := args[0].(int64)
		*buf = &a
	case *int64:
		a := *args[0].(*int64)
		*buf = &a
	case uint:
		a := args[0].(uint)
		*buf = &a
	case *uint:
		a := *args[0].(*uint)
		*buf = &a
	case uint8:
		a := args[0].(uint8)
		*buf = &a
	case *uint8:
		a := *args[0].(*uint8)
		*buf = &a
	case uint16:
		a := args[0].(uint16)
		*buf = &a
	case *uint16:
		a := *args[0].(*uint16)
		*buf = &a
	case uint32:
		a := args[0].(uint32)
		*buf = &a
	case *uint32:
		a := *args[0].(*uint32)
		*buf = &a
	case uint64:
		a := args[0].(uint64)
		*buf = &a
	case *uint64:
		a := *args[0].(*uint64)
		*buf = &a
	case float32:
		a := args[0].(float32)
		*buf = &a
	case *float32:
		a := *args[0].(*float32)
		*buf = &a
	case float64:
		a := args[0].(float64)
		*buf = &a
	case *float64:
		a := *args[0].(*float64)
		*buf = &a
	case *vector.Node:
		node := args[0].(*vector.Node)
		if node != nil {
			i := ctx.reserveBB()
			ctx.bufBB[i] = append(ctx.bufBB[i], node.Bytes()...)
			*buf = &ctx.bufBB[i]
		}
	default:
		*buf = nil
	}
	return
}

// Conditional assignment modifier.
func modIfThen(_ *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
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
func modIfThenElse(_ *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
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
func checkTrue(val interface{}) (r bool) {
	switch val.(type) {
	case *[]byte:
		a := *val.(*[]byte)
		r = bytes.Equal(a, bTrue)
	case []byte:
		a := val.([]byte)
		r = bytes.Equal(a, bTrue)
	case *string:
		a := *val.(*string)
		r = a == "true"
	case string:
		a := val.(string)
		r = a == "true"
	case *bool:
		r = *val.(*bool)
	case bool:
		r = val.(bool)
	case int:
		r = val.(int) == 1
	case *int:
		r = *val.(*int) == 1
	case int8:
		r = val.(int8) == 1
	case *int8:
		r = *val.(*int8) == 1
	case int16:
		r = val.(int16) == 1
	case *int16:
		r = *val.(*int16) == 1
	case int32:
		r = val.(int32) == 1
	case *int32:
		r = *val.(*int32) == 1
	case int64:
		r = val.(int64) == 1
	case *int64:
		r = *val.(*int64) == 1
	case uint:
		r = val.(uint) == 1
	case *uint:
		r = *val.(*uint) == 1
	case uint8:
		r = val.(uint8) == 1
	case *uint8:
		r = *val.(*uint8) == 1
	case uint16:
		r = val.(uint16) == 1
	case *uint16:
		r = *val.(*uint16) == 1
	case uint32:
		r = val.(uint32) == 1
	case *uint32:
		r = *val.(*uint32) == 1
	case uint64:
		r = val.(uint64) == 1
	case *uint64:
		r = *val.(*uint64) == 1
	case float32:
		r = val.(float32) == 1
	case *float32:
		r = *val.(*float32) == 1
	case float64:
		r = val.(float64) == 1
	case *float64:
		r = *val.(*float64) == 1
	case *vector.Node:
		node := val.(*vector.Node)
		if node != nil {
			r = node.Bool()
		}
	default:
		r = false
	}
	return
}
