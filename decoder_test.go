package decoder

import (
	"testing"

	"github.com/koykov/inspector/testobj"
)

func TestDecoder(t *testing.T) {
	t.Run("decoder0", func(t *testing.T) { testDecoder(t, "src", scenarioDec0) })
	t.Run("decoder1", func(t *testing.T) { testDecoder(t, "src", scenarioDec1) })
	t.Run("decoder2", func(t *testing.T) { testDecoder(t, "src", scenarioDec2) })
	// t.Run("decoder3", func(t *testing.T) { testDecoder(t, "srcNested", scenarioDec3) }) // check decoder_legacy project
	t.Run("decoder4", func(t *testing.T) { testDecoder(t, "src", scenarioDec4) })

	t.Run("loop_range", func(t *testing.T) { testDecoder(t, "src", scenarioNop) })
	t.Run("loop_counter", func(t *testing.T) { testDecoder(t, "src", scenarioLoop1) })

	t.Run("cond", func(t *testing.T) { testDecoder(t, "src", scenarioCond) })
	t.Run("cond_else", func(t *testing.T) { testDecoder(t, "src", scenarioCond1) })
	t.Run("condOK", func(t *testing.T) { testDecoder(t, "src", scenarioCondOK) })
	t.Run("condNotOK", func(t *testing.T) { testDecoder(t, "src", scenarioCondOK1) })
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
	// b.Run("decoder3", func(b *testing.B) { benchDecoder(b, "srcNested", scenarioDec3) }) // check decoder_legacy project
	b.Run("decoder4", func(b *testing.B) { benchDecoder(b, "src", scenarioDec4) })

	b.Run("loop0", func(b *testing.B) { benchDecoder(b, "src", scenarioNop) })
	b.Run("loop1", func(b *testing.B) { benchDecoder(b, "src", scenarioLoop1) })

	b.Run("cond", func(b *testing.B) { benchDecoder(b, "src", scenarioCond) })
	b.Run("cond_else", func(b *testing.B) { benchDecoder(b, "src", scenarioCond1) })

	b.Run("condOK", func(b *testing.B) { benchDecoder(b, "src", scenarioCondOK) })
	b.Run("condNotOK", func(b *testing.B) { benchDecoder(b, "src", scenarioCondOK1) })
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

func scenarioNop(t testing.TB, obj *testobj.TestObject) {
	_, _ = t, obj
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

// check decoder_legacy project
// func scenarioDec3(t testing.TB, obj *testobj.TestObject) {
// 	assertS(t, "Id", obj.Id, "xFF45")
// 	assertF64(t, "Cost", obj.Cost, 123)
// }

func scenarioDec4(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Name", obj.Name, []byte(`2677594116`))
}

func scenarioLoop1(t testing.TB, obj *testobj.TestObject) {
	assertI32(t, "Status", obj.Status, 66)
}

func scenarioCond(t testing.TB, obj *testobj.TestObject) {
	assertU64(t, "Ustate", obj.Ustate, 17)
}

func scenarioCond1(t testing.TB, obj *testobj.TestObject) {
	assertU64(t, "Ustate", obj.Ustate, 23)
}

func scenarioCondOK(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Id", obj.Id, "15")
}

func scenarioCondOK1(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Name", obj.Name, []byte("N/D"))
}
