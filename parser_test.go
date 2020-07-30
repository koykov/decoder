package jsondecoder

import (
	"bytes"
	"testing"
)

var (
	v2vEx0 = []byte(`dst.ID = obj.user_id
dst.Name = "Jonh Ruth"
dst.Finance.Balance = obj.cost.total`)
	v2vEx0Expect = []byte(`dst: dst.ID <- src: obj.user_id
dst: dst.Name <- src: "Jonh Ruth"
dst: dst.Finance.Balance <- src: obj.cost.total
`)
)

func TestParse_V2V(t *testing.T) {
	rules, err := Parse(v2vEx0)
	if err != nil {
		t.Error(err)
	}
	r := rules.HumanReadable()
	if !bytes.Equal(r, v2vEx0Expect) {
		t.Errorf("v2v example 0 test failed\nexp: %s\ngot: %s", v2vEx0Expect, r)
	}
}
