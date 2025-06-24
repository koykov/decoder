package decoder

import (
	"github.com/koykov/inspector"
	"github.com/koykov/vector"
)

// AssignVectorNode implements assign callback to convert vector.Node to destination with arbitrary type.
func AssignVectorNode(dst, src any, _ inspector.AccumulativeBuffer) (ok bool) {
	switch x := src.(type) {
	case *vector.Node:
		n := x
		if n.Type() == vector.TypeNull {
			return
		}
		ok = true
		switch y := dst.(type) {
		case *[]byte:
			*y = n.Bytes()
		case *string:
			*y = n.String()
		case *bool:
			*y = n.Bool()
		case *int:
			i, _ := n.Int()
			*y = int(i)
		case *int8:
			i, _ := n.Int()
			*y = int8(i)
		case *int16:
			i, _ := n.Int()
			*y = int16(i)
		case *int32:
			i, _ := n.Int()
			*y = int32(i)
		case *int64:
			i, _ := n.Int()
			*y = i
		case *uint:
			u, _ := n.Uint()
			*y = uint(u)
		case *uint8:
			u, _ := n.Uint()
			*y = uint8(u)
		case *uint16:
			u, _ := n.Uint()
			*y = uint16(u)
		case *uint32:
			u, _ := n.Uint()
			*y = uint32(u)
		case *uint64:
			u, _ := n.Uint()
			*y = u
		case *float32:
			f, _ := n.Float()
			*y = float32(f)
		case *float64:
			f, _ := n.Float()
			*y = f
		default:
			ok = false
		}
	}
	return
}
