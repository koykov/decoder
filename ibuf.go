package decoder

import "github.com/koykov/inspector"

type ibuf struct {
	cnt int
	buf []any
}

type ibufs struct {
	index map[string]int
	buf   []ibuf
}

func (b *ibufs) init() {
	if b.index == nil {
		b.index = make(map[string]int)
	}
}

func (b *ibufs) get(key string) (any, error) {
	b.init()

	ins, err := inspector.GetInspector(key)
	if err != nil {
		return nil, err
	}

	i, ok := b.index[key]
	if !ok {
		b.buf = append(b.buf, ibuf{})
		i = len(b.buf) - 1
		b.index[key] = i
	}

	ib := &b.buf[i]
	if ib.cnt < len(ib.buf) {
		val := ib.buf[ib.cnt]
		ib.cnt++
		return val, nil
	} else {
		ib.buf = append(ib.buf, ins.Instance(true))
		val := ib.buf[ib.cnt]
		ib.cnt++
		return val, nil
	}
}

func (b *ibufs) reset() {
	b.init()
	for iname, idx := range b.index {
		ins, err := inspector.GetInspector(iname)
		if err != nil {
			continue
		}
		ib := &b.buf[idx]
		for i := 0; i < ib.cnt; i++ {
			_ = ins.Reset(ib.buf[i])
		}
		ib.cnt = 0
	}
}
