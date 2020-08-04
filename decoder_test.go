package jsondecoder

import (
	"bytes"
	"testing"

	"github.com/koykov/inspector/testobj"
	"github.com/koykov/inspector/testobj_ins"
)

var (
	decTestSrc = []byte(`{
  "identifier": "xf44e",
  "person": {
    "full_name": "Marquis Warren",
    "status": 67,
    "read_f": 4,
    "write_f": 8,
	"last_buy": 45.90421
  },
  "finance": {
    "balance": 164.5962,
	"balance_total": 200,
    "is_active": true
  },
  "ext": {
    "perm": [false, true, true]
  }
}`)

	decTest0 = []byte(`obj.Id = jso.identifier
obj.Name = jso.person.full_name
obj.Status = jso.person.status
obj.Cost = jso.finance.balance
obj.Finance.AllowBuy = jso.finance.is_active
obj.Flags[read] = jso.person.read_f
obj.Flags[write] = jso.person.write_f
obj.Permission[45] = jso.ext.perm[0]
obj.Permission[59] = jso.ext.perm[2]`)
	decTest1 = []byte(`obj.Id = jso.identifier
obj.Name = jso.person.full_name
obj.Status = jso.person.status
obj.Cost = jso.finance.balance
obj.Finance.AllowBuy = jso.finance.is_active
obj.Finance.MoneyIn = 15.4532
obj.Finance.MoneyOut = jso.person.last_buy
obj.Finance.Balance = jso.finance.balance_total`)
)

func pretest(t testing.TB) {
	dec := map[string][]byte{
		"decTest0": decTest0,
		"decTest1": decTest1,
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
	err := ctx.SetJson("jso", decTestSrc)
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
	if obj.Flags["read"] != 4 {
		t.Error("decode 0 flags[read] test failed")
	}
	if obj.Flags["write"] != 8 {
		t.Error("decode 0 flags[write] test failed")
	}
	perm := obj.Permission
	if (*perm)[45] != false {
		t.Error("decode 0 permission[45] test failed")
	}
	if (*perm)[59] != true {
		t.Error("decode 0 permission[45] test failed")
	}
}

func assertTest1(t testing.TB, obj *testobj.TestObject) {
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
	if obj.Finance.MoneyIn != 15.4532 {
		t.Error("decode 0 finance.moneyIn test failed")
	}
	if obj.Finance.MoneyOut != 45.90421 {
		t.Error("decode 0 finance.moneyOut test failed")
	}
	if obj.Finance.Balance != 200 {
		t.Error("decode 0 finance.balance test failed")
	}
}

func TestDecode1(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	err := ctx.SetJson("jso", decTestSrc)
	if err != nil {
		t.Error(err)
	}
	err = Decode("decTest1", ctx)
	if err != nil {
		t.Error(err)
	}
	assertTest1(t, obj)
}

func BenchmarkDecode1(b *testing.B) {
	pretest(b)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
		err := ctx.SetJson("jso", decTestSrc)
		if err != nil {
			b.Error(err)
		}
		err = Decode("decTest1", ctx)
		if err != nil {
			b.Error(err)
		}
		assertTest1(b, obj)
		ctx.Reset()
		obj.Clear()
	}
}
