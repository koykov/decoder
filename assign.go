package decoder

import "github.com/koykov/jsonvector"

func AssignJsonNode(dst, src interface{}) (ok bool) {
	switch src.(type) {
	case *jsonvector.Node:
		n := src.(*jsonvector.Node)
		if n == nil {
			return
		}
		ok = true
		switch dst.(type) {
		case *[]byte:
			*dst.(*[]byte) = n.Bytes()
		case *string:
			*dst.(*string) = n.String()
		case *bool:
			*dst.(*bool) = n.Bool()
		case *int:
			*dst.(*int) = int(n.Int())
		case *int8:
			*dst.(*int8) = int8(n.Int())
		case *int16:
			*dst.(*int16) = int16(n.Int())
		case *int32:
			*dst.(*int32) = int32(n.Int())
		case *int64:
			*dst.(*int64) = n.Int()
		case *uint:
			*dst.(*uint) = uint(n.Uint())
		case *uint8:
			*dst.(*uint8) = uint8(n.Uint())
		case *uint16:
			*dst.(*uint16) = uint16(n.Uint())
		case *uint32:
			*dst.(*uint32) = uint32(n.Uint())
		case *uint64:
			*dst.(*uint64) = n.Uint()
		case *float32:
			*dst.(*float32) = float32(n.Float())
		case *float64:
			*dst.(*float64) = n.Float()
		default:
			ok = false
		}
	}
	return
}
