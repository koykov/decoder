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

	t.Run("loop_counter", testParser)
	t.Run("loop_range", testParser)
	t.Run("loop_break", testParser)
	t.Run("loop_lazybreak", testParser)
	t.Run("loop_continue", testParser)

	t.Run("cond", testParser)
	t.Run("cond_else", testParser)
	t.Run("cond_helper", testParser)
	t.Run("condOK", testParser)
	t.Run("condNotOK", testParser)

	t.Run("switch", testParser)
	t.Run("switch_no_cond", testParser)
}

func testParser(t *testing.T) {
	key := getTBName(t)
	st := getStage("parser/" + key)
	if st == nil {
		t.Error("stage not found")
		return
	}
	if len(st.expect) > 0 {
		rs, err := Parse(st.origin)
		if err != nil {
			t.Error(err)
		}
		r := rs.HumanReadable()
		if !bytes.Equal(r, st.expect) {
			t.Errorf("%s test failed\nexp: %s\ngot: %s", key, string(st.expect), string(r))
		}
	}
}
