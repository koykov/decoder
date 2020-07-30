package jsondecoder

import (
	"bytes"
	"testing"
)

var (
	v2vEx0 = []byte(`dst.ID = obj.user_id
dst.Name = "Jonh Ruth"
dst.Finance.Balance = obj.cost.total
dst.Weight = 12.45`)
	v2vEx0Expect = []byte(`dst: dst.ID <- src: obj.user_id
dst: dst.Name <- src: "Jonh Ruth"
dst: dst.Finance.Balance <- src: obj.cost.total
dst: dst.Weight <- src: "12.45"
`)
	v2vEx1 = []byte(`person.Gender = request.gender|default("male")
person.Owner = false`)
	v2vEx1Expect = []byte(`dst: person.Gender <- src: request.gender mod default("male")
dst: person.Owner <- src: "false"
`)
)

func TestParse_V2V(t *testing.T) {
	var (
		rules Rules
		err   error
		r     []byte
	)

	if rules, err = Parse(v2vEx0); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, v2vEx0Expect) {
		t.Errorf("v2v example 0 test failed\nexp: %s\ngot: %s", v2vEx0Expect, r)
	}

	rules, err = Parse(v2vEx1)
	if rules, err = Parse(v2vEx1); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, v2vEx1Expect) {
		t.Errorf("v2v example 1 test failed\nexp: %s\ngot: %s", v2vEx1Expect, r)
	}
}
