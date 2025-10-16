package decoder

import (
	"bytes"
	"testing"

	"github.com/koykov/inspector/testobj"
)

func TestModNewBufferize(t *testing.T) {
	testfn := func(t *testing.T, checkfn func(ctx *Ctx) bool) {
		ctx := NewCtx()
		key := "mod/" + getTBName(t)
		err := Decode(key, ctx)
		if err != nil {
			t.Error(err)
		}
		if !checkfn(ctx) {
			t.Fail()
		}
	}
	t.Run("new", func(t *testing.T) {
		testfn(t, func(ctx *Ctx) (r bool) {
			if obj := ctx.Get("t"); t != nil {
				r = bytes.Equal(obj.(*testobj.TestObject).Name, []byte("foobar"))
			}
			return
		})
	})
	t.Run("bufferize", func(t *testing.T) {
		testfn(t, func(ctx *Ctx) (r bool) {
			if hist := ctx.Get("h"); hist != nil {
				r = hist.(*testobj.TestHistory).Cost == 4
			}
			return
		})
	})
}

func BenchmarkModNewBufferize(b *testing.B) {
	benchfn := func(b *testing.B) {
		b.ReportAllocs()
		ctx := NewCtx()
		key := "mod/" + getTBName(b)
		for i := 0; i < b.N; i++ {
			_ = Decode(key, ctx)
			ctx.Reset()
		}
	}
	b.Run("new", func(b *testing.B) { benchfn(b) }) // this bench must have alloc!
	b.Run("bufferize", func(b *testing.B) { benchfn(b) })
}
