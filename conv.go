package decoder

import (
	"strconv"

	"github.com/koykov/bytebuf"
	"github.com/koykov/byteconv"
)

type intConverter interface {
	Int() (int64, error)
}

// Convert interface value with arbitrary underlying type to integer value.
func iface2int(raw any) (r int64, ok bool) {
	ok = true
	switch x := raw.(type) {
	case int:
		r = int64(x)
	case *int:
		r = int64(*x)
	case int8:
		r = int64(x)
	case *int8:
		r = int64(*x)
	case int16:
		r = int64(x)
	case *int16:
		r = int64(*x)
	case int32:
		r = int64(x)
	case *int32:
		r = int64(*x)
	case int64:
		r = x
	case *int64:
		r = *x
	case uint:
		r = int64(x)
	case *uint:
		r = int64(*x)
	case uint8:
		r = int64(x)
	case *uint8:
		r = int64(*x)
	case uint16:
		r = int64(x)
	case *uint16:
		r = int64(*x)
	case uint32:
		r = int64(x)
	case *uint32:
		r = int64(*x)
	case uint64:
		r = int64(x)
	case *uint64:
		r = int64(*x)
	case []byte:
		if len(x) > 0 {
			r, _ = strconv.ParseInt(byteconv.B2S(x), 0, 0)
		}
	case *[]byte:
		if len(*x) > 0 {
			r, _ = strconv.ParseInt(byteconv.B2S(*x), 0, 0)
		}
	case string:
		if len(x) > 0 {
			r, _ = strconv.ParseInt(x, 0, 0)
		}
	case *string:
		if len(*x) > 0 {
			r, _ = strconv.ParseInt(*x, 0, 0)
		}
	case *bytebuf.Chain:
		if x.Len() > 0 {
			r, _ = strconv.ParseInt(x.String(), 0, 0)
		}
	case *bytebuf.Accumulative:
		if x.Len() > 0 {
			r, _ = strconv.ParseInt(x.StakedString(), 0, 0)
		}
	case intConverter:
		if i, err := x.Int(); err == nil {
			return i, true
		}
	default:
		ok = false
	}
	return
}
