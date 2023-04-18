package decoder

import (
	"github.com/koykov/inspector"
	"github.com/koykov/vector"
)

// AssignVectorNode implements assign callback to convert vector.Node to destination with arbitrary type.
func AssignVectorNode(dst, src any, _ inspector.AccumulativeBuffer) (ok bool) {
	switch src.(type) {
	case *vector.Node:
		n := src.(*vector.Node)
		if n.Type() == vector.TypeNull {
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
			i, _ := n.Int()
			*dst.(*int) = int(i)
		case *int8:
			i, _ := n.Int()
			*dst.(*int8) = int8(i)
		case *int16:
			i, _ := n.Int()
			*dst.(*int16) = int16(i)
		case *int32:
			i, _ := n.Int()
			*dst.(*int32) = int32(i)
		case *int64:
			i, _ := n.Int()
			*dst.(*int64) = i
		case *uint:
			u, _ := n.Uint()
			*dst.(*uint) = uint(u)
		case *uint8:
			u, _ := n.Uint()
			*dst.(*uint8) = uint8(u)
		case *uint16:
			u, _ := n.Uint()
			*dst.(*uint16) = uint16(u)
		case *uint32:
			u, _ := n.Uint()
			*dst.(*uint32) = uint32(u)
		case *uint64:
			u, _ := n.Uint()
			*dst.(*uint64) = u
		case *float32:
			f, _ := n.Float()
			*dst.(*float32) = float32(f)
		case *float64:
			f, _ := n.Float()
			*dst.(*float64) = f
		default:
			ok = false
		}
	}
	return
}
