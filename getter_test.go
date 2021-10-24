package decoder

import (
	"testing"

	"github.com/koykov/inspector/testobj"
)

func TestGetter(t *testing.T) {
	t.Run("crc32", func(t *testing.T) { testGetter(t, "src", scenarioGetterCrc32) })
	t.Run("crc32Static", func(t *testing.T) { testGetter(t, "src", scenarioGetterCrc32) })
	t.Run("atoi", func(t *testing.T) { testGetter(t, "src", scenarioGetterAtoi) })
	t.Run("atou", func(t *testing.T) { testGetter(t, "src", scenarioGetterAtou) })
	t.Run("atof", func(t *testing.T) { testGetter(t, "src", scenarioGetterAtof) })
	t.Run("atob", func(t *testing.T) { testGetter(t, "src", scenarioGetterAtob) })
	t.Run("itoa-explicit-vector", func(t *testing.T) { testGetter(t, "src", scenarioGetterItoa) })
	t.Run("itoa-explicit-var", func(t *testing.T) { testGetter(t, "src", scenarioGetterItoa) })
	t.Run("itoa-implicit-vector", func(t *testing.T) { testGetter(t, "src", scenarioGetterItoa) })
	t.Run("itoa-implicit-var", func(t *testing.T) { testGetter(t, "src", scenarioGetterItoa) })
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
	b.Run("atoi", func(b *testing.B) { benchGetter(b, "src", scenarioGetterAtoi) })
	b.Run("atou", func(b *testing.B) { benchGetter(b, "src", scenarioGetterAtou) })
	b.Run("atof", func(b *testing.B) { benchGetter(b, "src", scenarioGetterAtof) })
	b.Run("atob", func(b *testing.B) { benchGetter(b, "src", scenarioGetterAtob) })
	b.Run("itoa-explicit-vector", func(b *testing.B) { benchGetter(b, "src", scenarioGetterItoa) })
	b.Run("itoa-explicit-var", func(b *testing.B) { benchGetter(b, "src", scenarioGetterItoa) })
	b.Run("itoa-implicit-vector", func(b *testing.B) { benchGetter(b, "src", scenarioGetterItoa) })
	b.Run("itoa-implicit-var", func(b *testing.B) { benchGetter(b, "src", scenarioGetterItoa) })
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

func scenarioGetterAtoi(t testing.TB, obj *testobj.TestObject) {
	assertI32(t, "Status", obj.Status, 105999)
}

func scenarioGetterAtou(t testing.TB, obj *testobj.TestObject) {
	assertU64(t, "Status", obj.Ustate, 67)
}

func scenarioGetterAtof(t testing.TB, obj *testobj.TestObject) {
	assertF64(t, "Cost", obj.Cost, 45.90421)
}

func scenarioGetterAtob(t testing.TB, obj *testobj.TestObject) {
	assertBl(t, "Finance.AllowBuy", obj.Finance.AllowBuy, true)
}

func scenarioGetterItoa(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Status", obj.Name, []byte("67"))
}
