package jsondecoder

import (
	"bytes"
	"testing"

	"github.com/koykov/inspector/testobj"
	"github.com/koykov/inspector/testobj_ins"
)

var (
	decTest0 = []byte(`obj.Id = jso.identifier
obj.Name = jso.person.full_name
obj.Status = jso.person.status
obj.Cost = jso.finance.balance
obj.Finance.AllowBuy = jso.finance.is_active
obj.Flags[read] = jso.person.read_f
obj.Flags[write] = jso.person.write_f
obj.Permission[45] = jso.ext.perm[0]
obj.Permission[59] = jso.ext.perm[2]`)
	decTestSrc0 = []byte(`{
  "identifier": "xf44e",
  "person": {
    "full_name": "Marquis Warren",
    "status": 67,
    "read_f": 4,
    "write_f": 8
  },
  "finance": {
    "balance": 164.5962,
    "is_active": true
  },
  "ext": {
    "perm": [45, 28, 59]
  }
}`)
)

func pretest(t testing.TB) {
	dec := map[string][]byte{
		"decTest0": decTest0,
	}
	for id, body := range dec {
		rules, err := Parse(body)
		if err != nil {
			t.Errorf("%s parse error %s", id, err)
		}
		RegisterDecoder(id, rules)
	}
}

func TestDecode0(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	err := ctx.SetJson("jso", decTestSrc0)
	if err != nil {
		t.Error(err)
	}
	err = Decode("decTest0", ctx)
	if err != nil {
		t.Error(err)
	}
	if obj.Id != "xf44e" {
		t.Error("decode 0 id test failed")
	}
	if !bytes.Equal(obj.Name, []byte("Marquis Warren")) {
		t.Error("decode 0 name test failed")
	}
	if obj.Status != 67 {
		t.Error("decode 0 status test failed")
	}
	if obj.Cost != 164.5962 {
		t.Error("decode 0 cost test failed")
	}
	if obj.Finance.AllowBuy != true {
		t.Error("decode 0 finance.allowBuy test failed")
	}
}