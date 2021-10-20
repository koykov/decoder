package decoder

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stage struct {
	key, err            string
	origin, expect, raw []byte
}

var (
	stages []stage
)

func init() {
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
