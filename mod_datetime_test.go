package decoder

import (
	"bytes"
	"testing"
	"time"
)

var (
	dt97, _ = time.Parse("2006-01-02", "1997-04-19")
	dt0     = time.Unix(1136239445, 123456789).UTC()

	dtRFC3339 = []time.Time{
		time.Date(2008, 9, 17, 20, 4, 26, 0, time.UTC),
		time.Date(1994, 9, 17, 20, 4, 26, 0, time.FixedZone("EST", -18000)),
		time.Date(2000, 12, 26, 1, 15, 6, 0, time.FixedZone("OTO", 15600)),
	}

	loc, _   = time.LoadLocation("Europe/Moscow")
	dtNative = time.Unix(0, 1233810057012345600).In(loc)
	dtTZ     = time.Date(1994, 9, 17, 20, 4, 26, 0, time.FixedZone("EST", -18000))
	dtAdd    = time.Date(2012, 1, 21, 20, 4, 26, 555, time.UTC)
)

func TestModDatetime(t *testing.T) {
	testfn := func(t *testing.T, dt time.Time) {
		key := "datetime/" + getTBName(t)
		st := getStage(key)

		lvalue, lvalue1, lvalue2, lvalue3 := make([]byte, 0), make([]byte, 0), make([]byte, 0), make([]byte, 0)
		ctx := NewCtx()
		ctx.SetStatic("lvalue", &lvalue)
		ctx.SetStatic("lvalue1", &lvalue1)
		ctx.SetStatic("lvalue2", &lvalue2)
		ctx.SetStatic("lvalue3", &lvalue3)
		ctx.SetStatic("date", &dt)
		err := Decode(key, ctx)
		if err != nil {
			t.Error(err)
		}
		expects := bytes.Split(st.expect, []byte("\n"))
		if !bytes.Equal(lvalue, expects[0]) {
			t.Errorf("got %s\nwant %s", lvalue, st.expect)
		}
		if len(expects) > 1 && !bytes.Equal(lvalue1, expects[1]) {
			t.Errorf("got %s\nwant %s", lvalue1, expects[1])
		}
		if len(expects) > 2 && !bytes.Equal(lvalue2, expects[2]) {
			t.Errorf("got %s\nwant %s", lvalue2, expects[2])
		}
		if len(expects) > 3 && !bytes.Equal(lvalue3, expects[3]) {
			t.Errorf("got %s\nwant %s", lvalue3, expects[3])
		}
	}

	t.Run("now", func(t *testing.T) { testfn(t, dt0) })

	t.Run("datePercent", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateYearShort", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateYear", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateMonth", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateMonthNameShort", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateMonthName", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateWeekNumberSun", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateWeekNumberMon", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateDay", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateDaySpacePad", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateDayOfYear", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateDayOfWeek", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateDayOfWeekISO", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateDayNameShort", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateDayName", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateHour", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateHourSpacePad", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateHour12", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateHour12SpacePad", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateMinute", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateSecond", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateAM_PM", func(t *testing.T) { testfn(t, dt0) })
	t.Run("date_am_pm", func(t *testing.T) { testfn(t, dt0) })
	t.Run("datePreferredTime", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateUnixtime", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateComplex_r", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateComplexR", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateComplexT", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateComplexc", func(t *testing.T) { testfn(t, dt0) })
	t.Run("dateComplexD", func(t *testing.T) { testfn(t, dt97) })
	t.Run("dateComplexF", func(t *testing.T) { testfn(t, dt97) })

	t.Run("dateRFC3339_0", func(t *testing.T) { testfn(t, dtRFC3339[0]) })
	t.Run("dateRFC3339_1", func(t *testing.T) { testfn(t, dtRFC3339[1]) })
	t.Run("dateRFC3339_2", func(t *testing.T) { testfn(t, dtRFC3339[2]) })

	t.Run("dateLayout", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateANSIC", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateUnixDate", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRubyDate", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRFC822", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRFC822Z", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRFC850", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRFC1123", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRFC1123Z", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRFC3339", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateRFC3339Nano", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateKitchen", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateStamp", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateStampMilli", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateStampMicro", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateStampNano", func(t *testing.T) { testfn(t, dtNative) })
	t.Run("dateLayoutTZ", func(t *testing.T) { testfn(t, dtTZ) })

	t.Run("addNS", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addUS", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addMS", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addS", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addM", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addH", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addD", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addW", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addMM", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addY", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addC", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addMIL", func(t *testing.T) { testfn(t, dtAdd) })
	t.Run("addMixed", func(t *testing.T) { testfn(t, dtAdd) })
}

func BenchmarkModDatetime(b *testing.B) {
	benchfn := func(b *testing.B, dt time.Time) {
		key := "datetime/" + getTBName(b)
		lvalue := make([]byte, 0)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ctx := AcquireCtx()
			ctx.SetStatic("lvalue", &lvalue)
			ctx.SetStatic("date", &dt)
			err := Decode(key, ctx)
			if err != nil {
				b.Error(err)
			}
			ReleaseCtx(ctx)
		}
	}
	b.Run("now", func(b *testing.B) { benchfn(b, dt0) })

	b.Run("dateComplexR", func(b *testing.B) { benchfn(b, dt0) })
	b.Run("dateStampNano", func(b *testing.B) { benchfn(b, dtNative) })

	b.Run("addMixedBench", func(b *testing.B) { benchfn(b, dtAdd) })
}
