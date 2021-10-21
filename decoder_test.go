package decoder

import (
	"testing"

	"github.com/koykov/inspector/testobj"
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
)

func TestDecoder(t *testing.T) {
	t.Run("decoder0", func(t *testing.T) { testDecoder(t, "src", scenario0) })
	t.Run("decoder1", func(t *testing.T) { testDecoder(t, "src", scenario1) })
	t.Run("decoder2", func(t *testing.T) { testDecoder(t, "src", scenario2) })
	t.Run("decoder3", func(t *testing.T) { testDecoder(t, "srcNested", scenario3) })
	t.Run("decoder4", func(t *testing.T) { testDecoder(t, "src", scenario4) })
}

func pretest(t testing.TB) {
	dec := map[string][]byte{
		"decModDefault0": decModDefault0,
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

func testDecoder(t *testing.T, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	obj = assertDecode(t, ctx, obj, jsonKey)
	assertFn(t, obj)
}

func BenchmarkDecoder(b *testing.B) {
	b.Run("decoder1", func(b *testing.B) { benchDecoder(b, "src", scenario1) })
	b.Run("decoder2", func(b *testing.B) { benchDecoder(b, "src", scenario2) })
	b.Run("decoder3", func(b *testing.B) { benchDecoder(b, "srcNested", scenario3) })
	b.Run("decoder4", func(b *testing.B) { benchDecoder(b, "src", scenario4) })
}

func benchDecoder(b *testing.B, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj = assertDecode(b, ctx, obj, jsonKey)
		assertFn(b, obj)
		obj.Clear()
	}
}

func scenario0(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Id", obj.Id, "xf44e")
	assertB(t, "Name", obj.Name, []byte("Marquis Warren"))
	assertI32(t, "Status", obj.Status, 67)
	assertF64(t, "Cost", obj.Cost, 164.5962)
	assertB1(t, "Finance.AllowBuy", obj.Finance.AllowBuy, true)
	assertI32(t, "Flags[read]", obj.Flags["read"], 4)
	assertI32(t, "Flags[write]", obj.Flags["write"], 8)
	perm := obj.Permission
	assertB1(t, "Permission[45]", (*perm)[45], false)
	assertB1(t, "Permission[59]", (*perm)[59], true)
}

func scenario1(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Id", obj.Id, "xf44e")
	assertB(t, "Name", obj.Name, []byte("Marquis Warren"))
	assertI32(t, "Status", obj.Status, 67)
	assertF64(t, "Cost", obj.Cost, 164.5962)
	assertB1(t, "Finance.AllowBuy", obj.Finance.AllowBuy, true)
	assertF64(t, "Finance.MoneyIn", obj.Finance.MoneyIn, 15.4532)
	assertF64(t, "Finance.MoneyOut", obj.Finance.MoneyOut, 45.90421)
	assertF64(t, "Finance.Balance", obj.Finance.Balance, 200)
}

func scenario2(t testing.TB, obj *testobj.TestObject) {
	assertI32(t, "len(Finance.History)", int32(len(obj.Finance.History)), 2)
	assertF64(t, "Finance.History[0].Cost", obj.Finance.History[0].Cost, 13.1415)
	assertF64(t, "Finance.History[1].Cost", obj.Finance.History[1].Cost, 164.5962)
}

func scenario3(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Id", obj.Id, "xFF45")
	assertF64(t, "Cost", obj.Cost, 123)
}

func scenario4(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Id", obj.Name, []byte(`2677594116`))
}
