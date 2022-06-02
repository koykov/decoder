package decoder

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/koykov/inspector/testobj"
	"github.com/koykov/inspector/testobj_ins"
)

type stage struct {
	key, err            string
	origin, expect, raw []byte
}

var (
	stages  []stage
	jsonSrc map[string][]byte
	buf     []byte
)

func init() {
	jsonSrc = make(map[string][]byte)

	_ = filepath.Walk("testdata/parser", func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".dec" {
			st := stage{}
			st.key = strings.Replace(filepath.Base(path), ".dec", "", 1)
			st.key = "parser/" + st.key

			st.origin, _ = ioutil.ReadFile(path)
			_, _ = Parse(st.origin)

			if raw, err := ioutil.ReadFile(strings.Replace(path, ".dec", ".xml", 1)); err == nil {
				st.expect = raw
			}
			stages = append(stages, st)
		}
		return nil
	})
	dirs := []string{"decoder", "mod", "getter"}
	for _, dir := range dirs {
		_ = filepath.Walk("testdata/"+dir, func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == ".dec" {
				st := stage{}
				st.key = strings.Replace(filepath.Base(path), ".dec", "", 1)
				st.key = dir + "/" + st.key

				st.origin, _ = ioutil.ReadFile(path)
				rules, _ := Parse(st.origin)
				RegisterDecoder(st.key, rules)

				stages = append(stages, st)
			}
			return nil
		})
	}
	_ = filepath.Walk("testdata/json", func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".json" {
			key := strings.Replace(filepath.Base(path), ".json", "", 1)
			src, _ := ioutil.ReadFile(path)
			jsonSrc[key] = src
		}
		return nil
	})
}

func getStage(key string) (st *stage) {
	for i := 0; i < len(stages); i++ {
		st1 := &stages[i]
		if st1.key == key {
			st = st1
		}
	}
	return st
}

func getTBName(tb testing.TB) string {
	key := tb.Name()
	return key[strings.Index(key, "/")+1:]
}

func assertDecode(t testing.TB, ctx *Ctx, obj *testobj.TestObject, target, jsonKey string) *testobj.TestObject {
	ctx.Reset()
	ctx.Set("obj", obj, &testobj_ins.TestObjectInspector{})
	buf = append(buf[:0], jsonSrc[jsonKey]...)
	_, err := ctx.SetVector("jso", buf, VectorJSON)
	if err != nil {
		t.Error(err)
	}
	ctx.SetStatic("ivar", int64(67))
	ctx.SetStatic("uvar", uint64(1e6))
	ctx.SetStatic("fvar", 3.1415)
	ctx.SetStatic("bvar", true)
	key := target + "/" + getTBName(t)
	err = Decode(key, ctx)
	if err != nil {
		t.Error(err)
	}
	return obj
}

func assertS(t testing.TB, field, val, expect string) {
	if val != expect {
		key := getTBName(t)
		t.Errorf("%s %s test failed", key, field)
	}
}

func assertB(t testing.TB, field string, val, expect []byte) {
	if !bytes.Equal(val, expect) {
		key := getTBName(t)
		t.Errorf("%s %s test failed", key, field)
	}
}

func assertI32(t testing.TB, field string, val, expect int32) {
	if val != expect {
		key := getTBName(t)
		t.Errorf("%s %s test failed", key, field)
	}
}

func assertU64(t testing.TB, field string, val, expect uint64) {
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

func assertBl(t testing.TB, field string, val, expect bool) {
	if val != expect {
		key := getTBName(t)
		t.Errorf("%s %s test failed", key, field)
	}
}
