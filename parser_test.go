package decoder

import (
	"bytes"
	"testing"
)

var (
	v2vEx0 = []byte(`dst.ID = obj.user_id
// example of comment
dst.Name = "John Ruth"
dst.Finance.Balance = obj.cost.total
dst.Weight = 12.45`)
	v2vEx0Expect = []byte(`dst: dst.ID <- src: obj.user_id
dst: dst.Name <- src: "John Ruth"
dst: dst.Finance.Balance <- src: obj.cost.total
dst: dst.Weight <- src: "12.45"
`)
	v2vEx1 = []byte(`person.Gender = request.gender|default("male")
person.Owner = false`)
	v2vEx1Expect = []byte(`dst: person.Gender <- src: request.gender mod default("male")
dst: person.Owner <- src: "false"
`)
	v2vEx2 = []byte(`dst.Id = src.id
dst.Status = src.{state|closed|expired}
dst.Hash = crc32("q", src.{id|title|descr})
foo(src.{a|b|c})`)
	v2vEx2Expect = []byte(`dst: dst.Id <- src: src.id
dst: dst.Status <- src: src.{state, closed, expired}
dst: dst.Hash <- src: crc32("q", src.{id, title, descr})
cb: foo(src.{a, b, c})
`)
	f2v = []byte(`bid.Id = 1
bid.Ext.HSum = crc32(response.title, response.val)
bid.Ext.Processed = response.Done|default(false)`)
	f2vExpect = []byte(`dst: bid.Id <- src: "1"
dst: bid.Ext.HSum <- src: crc32(response.title, response.val)
dst: bid.Ext.Processed <- src: response.Done mod default("false")
`)
	v2c = []byte(`dst.Hash = 64h95nd5fx
ctx.new = src.listing
dst.weight = new.value`)
	v2cExpect = []byte(`dst: dst.Hash <- src: 64h95nd5fx
dst: ctx.new <- src: src.listing
dst: dst.weight <- src: new.value
`)
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

	if rules, err = Parse(v2vEx1); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, v2vEx1Expect) {
		t.Errorf("v2v example 1 test failed\nexp: %s\ngot: %s", v2vEx1Expect, r)
	}

	if rules, err = Parse(v2vEx2); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, v2vEx2Expect) {
		t.Errorf("v2v example 2 test failed\nexp: %s\ngot: %s", v2vEx2Expect, r)
	}
}

func TestParse_F2V(t *testing.T) {
	var (
		rules Rules
		err   error
		r     []byte
	)

	if rules, err = Parse(f2v); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, f2vExpect) {
		t.Errorf("f2v example 0 test failed\nexp: %s\ngot: %s", f2vExpect, r)
	}
}

func TestParse_V2C(t *testing.T) {
	var (
		rules Rules
		err   error
		r     []byte
	)

	if rules, err = Parse(v2c); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, v2cExpect) {
		t.Errorf("v2c example 0 test failed\nexp: %s\ngot: %s", v2cExpect, r)
	}
}

func TestParse_V2CI(t *testing.T) {
	var (
		rules Rules
		err   error
		r     []byte
	)

	if rules, err = Parse(v2ci); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, v2ciExpect) {
		t.Errorf("v2ci example 0 test failed\nexp: %s\ngot: %s", v2ciExpect, r)
	}
}

func TestParse_Cb(t *testing.T) {
	var (
		rules Rules
		err   error
		r     []byte
	)

	if rules, err = Parse(cb); err != nil {
		t.Error(err)
	}
	r = rules.HumanReadable()
	if !bytes.Equal(r, cbExpect) {
		t.Errorf("callback example 0 test failed\nexp: %s\ngot: %s", cbExpect, r)
	}
}
