package decoder

import (
	"testing"

	"github.com/koykov/inspector/testobj"
)

func TestMod(t *testing.T) {
	t.Run("default", func(t *testing.T) { testMod(t, "src", scenarioModDefault) })
}

func testMod(t *testing.T, jsonKey string, assertFn func(t testing.TB, obj *testobj.TestObject)) {
	ctx := NewCtx()
	obj := &testobj.TestObject{}
	obj = assertDecode(t, ctx, obj, "mod", jsonKey)
	assertFn(t, obj)
}

func scenarioModDefault(t testing.TB, obj *testobj.TestObject) {
	assertB(t, "Name", obj.Name, []byte("Marquis Warren"))
	assertI32(t, "Status", obj.Status, 1)
}
