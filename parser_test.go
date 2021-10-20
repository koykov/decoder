package decoder

import (
	"bytes"
	"testing"
)

var (
	v2ci = []byte(`obj.Id = 17
ctx.finance = response.fin as Finance
obj.Balance = finance.Amount`)
	v2ciExpect = []byte(`dst: obj.Id <- src: "17"
dst: ctx.finance <- src: response.fin as Finance
dst: obj.Balance <- src: finance.Amount
`)
	cb = []byte(`user.Registered = 1
foo(user, req, true)
user.Status = 15`)
	cbExpect = []byte(`dst: user.Registered <- src: "1"
cb: foo(user, req, "true")
dst: user.Status <- src: "15"
`)
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

func TestParse_V2CI(t *testing.T) {
	var (
		rs  Ruleset
		err error
		r   []byte
	)

	if rs, err = Parse(v2ci); err != nil {
		t.Error(err)
	}
	r = rs.HumanReadable()
	if !bytes.Equal(r, v2ciExpect) {
		t.Errorf("v2ci example 0 test failed\nexp: %s\ngot: %s", v2ciExpect, r)
	}
}

func TestParse_Cb(t *testing.T) {
	var (
		rs  Ruleset
		err error
		r   []byte
	)

	if rs, err = Parse(cb); err != nil {
		t.Error(err)
	}
	r = rs.HumanReadable()
	if !bytes.Equal(r, cbExpect) {
		t.Errorf("callback example 0 test failed\nexp: %s\ngot: %s", cbExpect, r)
	}
}
