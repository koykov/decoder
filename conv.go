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
	switch raw.(type) {
	case int:
		r = int64(raw.(int))
	case *int:
		r = int64(*raw.(*int))
	case int8:
		r = int64(raw.(int8))
	case *int8:
		r = int64(*raw.(*int8))
	case int16:
		r = int64(raw.(int16))
	case *int16:
		r = int64(*raw.(*int16))
	case int32:
		r = int64(raw.(int32))
	case *int32:
		r = int64(*raw.(*int32))
	case int64:
		r = raw.(int64)
	case *int64:
		r = *raw.(*int64)
	case uint:
		r = int64(raw.(uint))
	case *uint:
		r = int64(*raw.(*uint))
	case uint8:
		r = int64(raw.(uint8))
	case *uint8:
		r = int64(*raw.(*uint8))
	case uint16:
		r = int64(raw.(uint16))
	case *uint16:
		r = int64(*raw.(*uint16))
	case uint32:
		r = int64(raw.(uint32))
	case *uint32:
		r = int64(*raw.(*uint32))
	case uint64:
		r = int64(raw.(uint64))
	case *uint64:
		r = int64(*raw.(*uint64))
	case []byte:
		if len(raw.([]byte)) > 0 {
			r, _ = strconv.ParseInt(byteconv.B2S(raw.([]byte)), 0, 0)
		}
	case *[]byte:
		if len(*raw.(*[]byte)) > 0 {
			r, _ = strconv.ParseInt(byteconv.B2S(*raw.(*[]byte)), 0, 0)
		}
	case string:
		if len(raw.(string)) > 0 {
			r, _ = strconv.ParseInt(raw.(string), 0, 0)
		}
	case *string:
		if len(*raw.(*string)) > 0 {
			r, _ = strconv.ParseInt(*raw.(*string), 0, 0)
		}
	case *bytebuf.Chain:
		if (*raw.(*bytebuf.Chain)).Len() > 0 {
			r, _ = strconv.ParseInt((*raw.(*bytebuf.Chain)).String(), 0, 0)
		}
	case intConverter:
		if i, err := raw.(intConverter).Int(); err == nil {
			return i, true
		}
	default:
		ok = false
	}
	return
}
