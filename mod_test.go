package decoder

import (
	"bytes"
	"testing"

	"github.com/koykov/inspector/testobj"
	"github.com/koykov/inspector/testobj_ins"
)

var (
	decModDefault0 = []byte(`obj.Name = jso.person.name|default(jso.person.full_name)
obj.Status = jso.person.state|default(1)`)
)

func TestModDefault(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	_, err := ctx.SetJson("jso", decTestSrc)
	if err != nil {
		t.Error(err)
	}
	err = Decode("decModDefault0", ctx)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(obj.Name, []byte("Marquis Warren")) {
		t.Error("decode mod default 0 name test failed")
	}
	if obj.Status != 1 {
		t.Error("decode mod default 0 status test failed")
	}
}
