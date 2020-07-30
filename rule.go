package jsondecoder

import "bytes"

type Rules []rule

type rule struct {
	dst, src []byte
	static   bool
}

func (r *Rules) HumanReadable() []byte {
	if len(*r) == 0 {
		return nil
	}
	var buf bytes.Buffer
	r.hrHelper(&buf)
	return buf.Bytes()
}

func (r *Rules) hrHelper(buf *bytes.Buffer) {
	for _, rule := range *r {
		buf.WriteString("dst: ")
		buf.Write(rule.dst)
		buf.WriteString(" <- src: ")
		if rule.static {
			buf.WriteByte('"')
			buf.Write(rule.src)
			buf.WriteByte('"')
		} else {
			buf.Write(rule.src)
		}
		buf.WriteByte('\n')
	}
}
