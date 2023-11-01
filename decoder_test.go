package decoder

import (
	"testing"

	"github.com/koykov/inspector/testobj"
)

func TestDecoder(t *testing.T) {
	t.Run("decoder0", func(t *testing.T) { testDecoder(t, "src", scenarioDec0) })
	t.Run("decoder1", func(t *testing.T) { testDecoder(t, "src", scenarioDec1) })
	t.Run("decoder2", func(t *testing.T) { testDecoder(t, "src", scenarioDec2) })
	// t.Run("decoder3", func(t *testing.T) { testDecoder(t, "srcNested", scenarioDec3) }) // check legacy package
	t.Run("decoder4", func(t *testing.T) { testDecoder(t, "src", scenarioDec4) })
}

func testDecoder(t *testing.T, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	obj = assertDecode(t, ctx, obj, "decoder", jsonKey)
	assertFn(t, obj)
}

func BenchmarkDecoder(b *testing.B) {
	b.Run("decoder1", func(b *testing.B) { benchDecoder(b, "src", scenarioDec1) })
	b.Run("decoder2", func(b *testing.B) { benchDecoder(b, "src", scenarioDec2) })
	// b.Run("decoder3", func(b *testing.B) { benchDecoder(b, "srcNested", scenarioDec3) }) // check legacy package
	b.Run("decoder4", func(b *testing.B) { benchDecoder(b, "src", scenarioDec4) })
}

func benchDecoder(b *testing.B, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj = assertDecode(b, ctx, obj, "decoder", jsonKey)
		assertFn(b, obj)
		obj.Clear()
	}
}

func scenarioDec0(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Id", obj.Id, "xf44e")
	assertB(t, "Name", obj.Name, []byte("Marquis Warren"))
	assertI32(t, "Status", obj.Status, 67)
	assertF64(t, "Cost", obj.Cost, 164.5962)
	assertBl(t, "Finance.AllowBuy", obj.Finance.AllowBuy, true)
	assertI32(t, "Flags[read]", obj.Flags["read"], 4)
	assertI32(t, "Flags[write]", obj.Flags["write"], 8)
	perm := obj.Permission
	assertBl(t, "Permission[45]", (*perm)[45], false)
	assertBl(t, "Permission[59]", (*perm)[59], true)
}

func scenarioDec1(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Id", obj.Id, "xf44e")
	assertB(t, "Name", obj.Name, []byte("Marquis Warren"))
	assertI32(t, "Status", obj.Status, 67)
	assertF64(t, "Cost", obj.Cost, 164.5962)
	assertBl(t, "Finance.AllowBuy", obj.Finance.AllowBuy, true)
	assertF64(t, "Finance.MoneyIn", obj.Finance.MoneyIn, 15.4532)
	assertF64(t, "Finance.MoneyOut", obj.Finance.MoneyOut, 45.90421)
	assertF64(t, "Finance.Balance", obj.Finance.Balance, 200)
}

func scenarioDec2(t testing.TB, obj *testobj.TestObject) {
	assertI32(t, "len(Finance.History)", int32(len(obj.Finance.History)), 2)
	assertF64(t, "Finance.History[0].Cost", obj.Finance.History[0].Cost, 13.1415)
	assertF64(t, "Finance.History[1].Cost", obj.Finance.History[1].Cost, 164.5962)
}

// check legacy package
// func scenarioDec3(t testing.TB, obj *testobj.TestObject) {
// 	assertS(t, "Id", obj.Id, "xFF45")
// 	assertF64(t, "Cost", obj.Cost, 123)
// }

func scenarioDec4(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Name", obj.Name, []byte(`2677594116`))
}
