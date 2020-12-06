package decoder

// Replace empty val with default value.
func modDefault(_ *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if val != nil {
		return
	}
	if len(args) == 0 {
		err = ErrModPoorArgs
	}
	switch args[0].(type) {
	case *[]byte:
		a := *args[0].(*byte)
		*buf = &a
	case []byte:
		a := args[0].(byte)
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
	default:
		*buf = nil
	}
	return
}
