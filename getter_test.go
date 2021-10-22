package decoder

import (
	"testing"

	"github.com/koykov/inspector/testobj"
)

func TestGetter(t *testing.T) {
	t.Run("crc32", func(t *testing.T) { testGetter(t, "src", scenarioGetterCrc32) })
	t.Run("crc32Static", func(t *testing.T) { testGetter(t, "src", scenarioGetterCrc32) })
}

func testGetter(t *testing.T, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	obj = assertDecode(t, ctx, obj, "getter", jsonKey)
	assertFn(t, obj)
}

func BenchmarkGetter(b *testing.B) {
	b.Run("crc32", func(b *testing.B) { benchGetter(b, "src", scenarioGetterCrc32) })
	b.Run("crc32Static", func(b *testing.B) { benchGetter(b, "src", scenarioGetterCrc32) })
}

func benchGetter(b *testing.B, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj = assertDecode(b, ctx, obj, "getter", jsonKey)
		assertFn(b, obj)
		obj.Clear()
	}
}

func scenarioGetterCrc32(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Name", obj.Id, "312073870")
}
