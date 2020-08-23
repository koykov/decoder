package decoder

import (
	"bytes"
	"testing"

	"github.com/koykov/inspector/testobj"
	"github.com/koykov/inspector/testobj_ins"
)

var (
	buf []byte

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
	decTestNestedJsonSrc = []byte(`{
	"id":"xFF45",
	"nickname":"Chris Mannix",
	"prop":"{\"id\":1,\"name\":\"Foo\",\"price\":123,\"tags\":[\"Bar\",\"Eek\"],\"stock\":{\"warehouse\":300,\"retail\":20}}"
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
	decTest2 = []byte(`obj.Finance.History = appendTestHistory(obj.Finance.History, 13.1415, jso.person.full_name)
obj.Finance.History = appendTestHistory(obj.Finance.History, jso.finance.balance, "foobar")`)
	decTest3 = []byte(`obj.Id = jso.id
obj.Name = jso.nickname
jsonParseAs(jso.prop, "properties")
obj.Cost = properties.price`)
	decTest4    = []byte(`obj.Name = crc32(jso.person.{id|nickname|full_name})`)
	crc32Expect = []byte(`2677594116`)
)

func pretest(t testing.TB) {
	dec := map[string][]byte{
		"decTest0": decTest0,
		"decTest1": decTest1,
		"decTest2": decTest2,
		"decTest3": decTest3,
		"decTest4": decTest4,
	}
	for id, body := range dec {
		rules, err := Parse(body)
		if err != nil {
			t.Errorf("%s parse error %s", id, err)
			continue
		}
		RegisterDecoder(id, rules)
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

func assertTest2(t testing.TB, obj *testobj.TestObject) {
	if len(obj.Finance.History) != 2 {
		t.Error("decode 2 history len mismatch")
	}
	if obj.Finance.History[0].Cost != 13.1415 {
		t.Error("decode 2 history row 0 cost mismatch")
	}
	if obj.Finance.History[1].Cost != 164.5962 {
		t.Error("decode 2 history row 1 cost mismatch")
	}
}

func assertTest3(t testing.TB, obj *testobj.TestObject) {
	if obj.Id != "xFF45" {
		t.Error("decode 3 id test failed")
	}
	if obj.Cost != 123 {
		t.Error("decode 3 cost mismatch")
	}
}

func assertTest4(t testing.TB, obj *testobj.TestObject) {
	if !bytes.Equal(obj.Name, crc32Expect) {
		t.Error("decode 4 id test failed")
	}
}

func TestDecode0(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	_, err := ctx.SetJson("jso", decTestSrc)
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

func TestDecode1(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	_, err := ctx.SetJson("jso", decTestSrc)
	if err != nil {
		t.Error(err)
	}
	err = Decode("decTest1", ctx)
	if err != nil {
		t.Error(err)
	}
	assertTest1(t, obj)
}

func TestDecode2(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	_, err := ctx.SetJson("jso", decTestSrc)
	if err != nil {
		t.Error(err)
	}
	err = Decode("decTest2", ctx)
	if err != nil {
		t.Error(err)
	}
	assertTest2(t, obj)
}

func TestDecode3(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	_, err := ctx.SetJson("jso", decTestNestedJsonSrc)
	if err != nil {
		t.Error(err)
	}
	err = Decode("decTest3", ctx)
	if err != nil {
		t.Error(err)
	}
	assertTest3(t, obj)
}

func TestDecode4(t *testing.T) {
	pretest(t)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	_, err := ctx.SetJson("jso", decTestSrc)
	if err != nil {
		t.Error(err)
	}
	err = Decode("decTest4", ctx)
	if err != nil {
		t.Error(err)
	}
	assertTest4(t, obj)
}

func BenchmarkDecode1(b *testing.B) {
	pretest(b)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
		_, err := ctx.SetJson("jso", decTestSrc)
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

func BenchmarkDecode2(b *testing.B) {
	pretest(b)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
		_, err := ctx.SetJson("jso", decTestSrc)
		if err != nil {
			b.Error(err)
		}
		err = Decode("decTest2", ctx)
		if err != nil {
			b.Error(err)
		}
		assertTest2(b, obj)
		ctx.Reset()
		obj.Clear()
	}
}

func BenchmarkDecode3(b *testing.B) {
	pretest(b)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf = append(buf[:0], decTestNestedJsonSrc...)

		ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
		_, err := ctx.SetJson("jso", buf)
		if err != nil {
			b.Error(err)
		}
		err = Decode("decTest3", ctx)
		if err != nil {
			b.Error(err)
		}
		assertTest3(b, obj)
		ctx.Reset()
		obj.Clear()
	}
}

func BenchmarkDecode4(b *testing.B) {
	pretest(b)
	obj := &testobj.TestObject{}
	ctx := NewCtx()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf = append(buf[:0], decTestSrc...)

		ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
		_, err := ctx.SetJson("jso", buf)
		if err != nil {
			b.Error(err)
		}
		err = Decode("decTest4", ctx)
		if err != nil {
			b.Error(err)
		}
		assertTest4(b, obj)
		ctx.Reset()
		obj.Clear()
	}
}
