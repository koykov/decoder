package decoder

import (
	"testing"

	"github.com/koykov/inspector/testobj"
)

func TestMod(t *testing.T) {
	t.Run("default", func(t *testing.T) { testMod(t, "src", scenarioModDefault) })
	t.Run("ifThen", func(t *testing.T) { testMod(t, "src", scenarioModIfThenElse) })
	t.Run("ifThenElse", func(t *testing.T) { testMod(t, "src", scenarioModIfThenElse) })
	t.Run("append", func(t *testing.T) { testMod(t, "src", scenarioModAppend) })
	t.Run("reset", func(t *testing.T) { testMod3(t, "src", scenarioModReset) })
}

func testMod(t *testing.T, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	obj = assertDecode(t, ctx, obj, "mod", jsonKey)
	assertFn(t, obj)
}

func testMod3(t *testing.T, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{Finance: &testobj.TestFinance{History: []testobj.TestHistory{
		{DateUnix: 111, Cost: 111},
		{DateUnix: 222, Cost: 222},
		{DateUnix: 333, Cost: 333},
	}}}
	obj = assertDecode(t, ctx, obj, "mod", jsonKey)
	assertFn(t, obj)
}

func BenchmarkMod(b *testing.B) {
	b.Run("default", func(b *testing.B) { benchMod(b, "src", scenarioModDefault, false) })
	b.Run("ifThen", func(b *testing.B) { benchMod(b, "src", scenarioModIfThenElse, false) })
	b.Run("ifThenElse", func(b *testing.B) { benchMod(b, "src", scenarioModIfThenElse, false) })
	b.Run("append", func(b *testing.B) { benchMod(b, "src", scenarioModAppend, false) })
	b.Run("reset", func(b *testing.B) { benchMod3(b, "src", scenarioModReset, true) })
}

func benchMod(b *testing.B, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject), noClear bool) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj = assertDecode(b, ctx, obj, "mod", jsonKey)
		assertFn(b, obj)
		if !noClear {
			obj.Clear()
		}
	}
}

func benchMod3(b *testing.B, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject), noClear bool) {
	ctx := NewCtx()
	obj := &testobj.TestObject{Finance: &testobj.TestFinance{History: []testobj.TestHistory{
		{DateUnix: 111, Cost: 111},
		{DateUnix: 222, Cost: 222},
		{DateUnix: 333, Cost: 333},
	}}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj = assertDecode(b, ctx, obj, "mod", jsonKey)
		assertFn(b, obj)
		if !noClear {
			obj.Clear()
		}
	}
}

func scenarioModDefault(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Name", obj.Name, []byte("Marquis Warren"))
	assertI32(t, "Status", obj.Status, 1)
}

func scenarioModIfThenElse(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Name", obj.Name, []byte("Rich men"))
}

func scenarioModAppend(t testing.TB, obj *testobj.TestObject) {
	if obj.Finance == nil {
		t.FailNow()
	}
	if len(obj.Finance.History) != 1 {
		t.FailNow()
	}
	x := obj.Finance.History[0]
	assertU64(t, "DateUnix", uint64(x.DateUnix), 111)
	assertB(t, "Comment", x.Comment, []byte("foobar"))
}

func scenarioModReset(t testing.TB, obj *testobj.TestObject) {
	if obj.Finance == nil {
		t.FailNow()
	}
	if len(obj.Finance.History) != 3 {
		t.FailNow()
	}
	x := obj.Finance.History[0]
	assertF64(t, "Cost", x.Cost, 0)
}
