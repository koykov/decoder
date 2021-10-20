package decoder

import (
	"bytes"
	"testing"
)

func TestParserV2V(t *testing.T) {
	t.Run("v2v0", func(t *testing.T) { testParser(t) })
	t.Run("v2v1", func(t *testing.T) { testParser(t) })
	t.Run("v2v2", func(t *testing.T) { testParser(t) })
}

func TestParserF2V(t *testing.T) {
	t.Run("f2v0", func(t *testing.T) { testParser(t) })
}

func TestParserV2C(t *testing.T) {
	t.Run("v2c0", func(t *testing.T) { testParser(t) })
}

func TestParserV2CI(t *testing.T) {
	t.Run("v2ci0", func(t *testing.T) { testParser(t) })
}

func TestParserCb(t *testing.T) {
	t.Run("cb0", func(t *testing.T) { testParser(t) })
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
