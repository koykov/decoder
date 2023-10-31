package decoder

import (
	"bytes"
	"testing"
)

func TestParser(t *testing.T) {
	t.Run("v2v0", testParser)
	t.Run("v2v1", testParser)
	t.Run("v2v2", testParser)
	t.Run("f2v0", testParser)
	t.Run("v2c0", testParser)
	t.Run("v2ci0", testParser)
	t.Run("cb0", testParser)
}

func testParser(t *testing.T) {
	key := getTBName(t)
	st := getStage("parser/" + key)
	if st == nil {
		t.Error("stage not found")
		return
	}
	if len(st.expect) > 0 {
		rs, _ := Parse(st.origin)
		r := rs.HumanReadable()
		if !bytes.Equal(r, st.expect) {
			t.Errorf("%s test failed\nexp: %s\ngot: %s", key, string(st.expect), string(r))
		}
	}
}
