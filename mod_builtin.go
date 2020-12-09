package decoder

import (
	"github.com/koykov/jsonvector"
)

// Replace empty val with default value.
func modDefault(_ *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
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
	case *jsonvector.Node:
		node := val.(*jsonvector.Node)
		empty = node == nil || node.Len() == 0
	default:
		empty = false
	}
	if !empty {
		return
	}
	if len(args) == 0 {
		err = ErrModPoorArgs
	}
	switch args[0].(type) {
	case *[]byte:
		a := *args[0].(*[]byte)
		*buf = &a
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
	case *jsonvector.Node:
		node := args[0].(*jsonvector.Node)
		if node != nil {
			a := node.Bytes()
			*buf = &a
		}
	default:
		*buf = nil
	}
	return
}
