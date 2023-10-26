package legacy

import (
	"strings"
	"testing"

	"github.com/koykov/decoder"
	"github.com/koykov/inspector/testobj"
	"github.com/koykov/inspector/testobj_ins"
)

const rulesSrc = `obj.Id = jso.id
obj.Name = jso.nickname
jsonParseAs(jso.prop, "properties")
obj.Cost = properties.price
`

var (
	src = []byte(`{
  "id":"xFF45",
  "nickname":"Chris Mannix",
  "prop":"{\"id\":1,\"name\":\"Foo\",\"price\":123,\"tags\":[\"Bar\",\"Eek\"],\"stock\":{\"warehouse\":300,\"retail\":20}}"
}
`)
	buf []byte
)

func init() {
	rules, _ := decoder.Parse([]byte(rulesSrc))
	decoder.RegisterDecoder("nestedJSON", rules)
}

func TestLegacy(t *testing.T) {
	ctx := decoder.NewCtx()
	obj := &testobj.TestObject{}
	obj = assertDecode(t, ctx, obj)
	scenarioDec3(t, obj)
}

func BenchmarkLegacy(b *testing.B) {
	ctx := decoder.NewCtx()
	obj := &testobj.TestObject{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj = assertDecode(b, ctx, obj)
		scenarioDec3(b, obj)
		obj.Clear()
	}
}

func assertDecode(t testing.TB, ctx *decoder.Ctx, obj *testobj.TestObject) *testobj.TestObject {
	ctx.Reset()
	ctx.Set("obj", obj, testobj_ins.TestObjectInspector{})
	buf = append(buf[:0], src...)
	_, err := ctx.SetVector("jso", buf, decoder.VectorJSON)
	if err != nil {
		t.Error(err)
	}
	ctx.SetStatic("ivar", int64(67))
	ctx.SetStatic("uvar", uint64(1e6))
	ctx.SetStatic("fvar", 3.1415)
	ctx.SetStatic("bvar", true)
	err = decoder.Decode("nestedJSON", ctx)
	if err != nil {
		t.Error(err)
	}
	return obj
}

func getTBName(tb testing.TB) string {
	key := tb.Name()
	return key[strings.Index(key, "/")+1:]
}

func scenarioDec3(t testing.TB, obj *testobj.TestObject) {
	assertS(t, "Id", obj.Id, "xFF45")
	assertF64(t, "Cost", obj.Cost, 123)
}

func assertS(t testing.TB, field, val, expect string) {
	if val != expect {
		key := getTBName(t)
		t.Errorf("%s %s test failed", key, field)
	}
}

func assertF64(t testing.TB, field string, val, expect float64) {
	if val != expect {
		key := getTBName(t)
		t.Errorf("%s %s test failed", key, field)
	}
}
