package decoder

import (
	"github.com/koykov/clock"
	"github.com/koykov/inspector"
)

func init() {
	// Register builtin modifiers.
	RegisterModFn("default", "def", modDefault).
		WithParam("arg any", "").
		WithDescription("Modifier `default` returns the passed `arg` if the preceding value is undefined or empty, otherwise the value of the variable.").
		WithExample(`obj.Name = jso.person.name|default(jso.person.full_name)
obj.Status = jso.person.state|default(1)`)
	RegisterModFn("ifThen", "if", modIfThen).
		WithDescription("Modifier `ifThen` passes `arg` only if preceding condition is true.").
		WithParam("arg any", "").
		WithExample(`obj.Name = jso.finance.is_active|ifThen("Rich men")`)
	RegisterModFn("ifThenElse", "ifel", modIfThenElse).
		WithDescription("Modifier `ifTheElse` passes `arg0` if preceding condition is true or `arg1` otherwise.").
		WithParam("arg0 any", "").
		WithParam("arg1 any", "").
		WithExample(`obj.Name = jso.finance.is_active|ifThenElse("Rich men", "Poor men")`)
	RegisterModFnNS("bar", "baz", "", func(_ *Ctx, _ *any, _ any, _ []any) error { return nil }).
		WithDescription("Testing stuff: don't use in production.")

	// Register fmt modifiers.
	RegisterModFnNS("fmt", "format", "f", modFmtFormat).
		WithDescription("Modifier `fmt::format` formats according to a format specifier and returns the resulting string.").
		WithParam("format string", "").
		WithParam("args ...any", "").
		WithExample("obj.StringField = fmt::format(\"Welcome %s\", user.Name)")

	// Register time modifiers.
	RegisterModFnNS("time", "now", "", modNow).
		WithDescription("Returns the current local time.")
	RegisterModFnNS("time", "format", "date", modDate).
		WithParam("layout string", "See https://github.com/koykov/clock#format for possible patterns").
		WithDescription("Modifier `time::format` returns a textual representation of the time value formatted according given layout.").
		WithExample(`lvalue = date|time::date("%d %b %y %H:%M %z") // 05 Feb 09 07:00 +0200
lvalue = date|time::date("%b %e %H:%M:%S.%N") // Feb  5 07:00:57.012345600`)
	RegisterModFnNS("time", "add", "date_modify", modDateAdd).
		WithParam("duration string", "Textual representation of duration you want to add to the datetime. Possible units:\n"+
			"  * `nsec`, `ns`\n"+
			"  * `usec`, `us`, `µs`\n"+
			"  * `msec`, `ms`\n"+
			"  * `seconds`, `second`, `sec`, `s`\n"+
			"  * `minutes`, `minute`, `min`, `m`\n"+
			"  * `hours`, `hour`, `hr`, `h`\n"+
			"  * `days`, `day`, `d`\n"+
			"  * `weeks`, `week`, `w`\n"+
			"  * `months`, `month`, `M`\n"+
			"  * `years`, `year`, `y`\n"+
			"  * `century`, `cen`, `c`\n"+
			"  * `millennium`, `mil`\n").
		WithDescription("Modifier `time::add` returns time+duration.").
		WithExample(`lvalue = date|time::add("+1 m")|time::date(time::StampNano)	// Jan 21 20:05:26.000000555
lvalue = date|time::add("+1 min")|time::date(time::StampNano)		// Jan 21 20:05:26.000000555
lvalue = date|time::add("+1 minute")|time::date(time::StampNano)	// Jan 21 20:05:26.000000555
lvalue = date|time::add("+1 minutes")|time::date(time::StampNano)	// Jan 21 20:05:26.000000555
`)

	// Register builtin getter callbacks.
	RegisterGetterFn("crc32", "", getterCrc32).
		WithParam("args ...any", "Arguments to concatenate.").
		WithDescription("Concatenate `args` and calculate crc32 IEEE checksum of result.")
	RegisterGetterFn("strToInt", "atoi", getterAtoi).
		WithParam("arg string", "Argument to convert.").
		WithDescription("Convert `arg` to `int` value if possible.")
	RegisterGetterFn("strToUint", "atou", getterAtou).
		WithParam("arg string", "Argument to convert.").
		WithDescription("Convert `arg` to `unsigned int` value if possible.")
	RegisterGetterFn("strToFloat", "atof", getterAtof).
		WithParam("arg string", "Argument to convert.").
		WithDescription("Convert `arg` to `float` value if possible.")
	RegisterGetterFn("strToBool", "atob", getterAtob).
		WithParam("arg string", "Argument to convert.").
		WithDescription("Convert `arg` to `bool` value if possible.")
	RegisterGetterFn("intToStr", "itoa", getterItoa).
		WithParam("arg int", "Argument to convert.").
		WithDescription("Convert `arg` to string value.")
	RegisterGetterFn("uintToStr", "utoa", getterUtoa).
		WithParam("arg int", "Argument to convert.").
		WithDescription("Convert `arg` to string value.")
	RegisterGetterFn("appendTestHistory", "", getterAppendTestHistory).
		WithDescription("Testing stuff: don't use in production.")

	// Register builtin callbacks.
	RegisterCallbackFnNS("fmt", "print", "", cbPrint).
		WithParam("args ...any", "Arguments to print.").
		WithDescription("Print args to console.")
	RegisterCallbackFnNS("fmt", "println", "", cbPrintln).
		WithParam("args ...any", "Arguments to print.").
		WithDescription("Print args to console with trailing newline.")

	// Register assign functions.
	inspector.RegisterAssignFn(AssignVectorNode)

	// Register datetime layouts.
	RegisterGlobalNS("time", "Layout", "", clock.Layout).
		WithType("string").
		WithDescription("time.Layout presentation in strtotime format.")
	RegisterGlobalNS("time", "ANSIC", "", clock.ANSIC).
		WithType("string").
		WithDescription("time.ANSIC presentation in strtotime format.")
	RegisterGlobalNS("time", "UnixDate", "", clock.UnixDate).
		WithType("string").
		WithDescription("time.UnixDate presentation in strtotime format.")
	RegisterGlobalNS("time", "RubyDate", "", clock.RubyDate).
		WithType("string").
		WithDescription("time.RubyDate presentation in strtotime format.")
	RegisterGlobalNS("time", "RFC822", "", clock.RFC822).
		WithType("string").
		WithDescription("time.RFC822 presentation in strtotime format.")
	RegisterGlobalNS("time", "RFC822Z", "", clock.RFC822Z).
		WithType("string").
		WithDescription("time.RFC822Z presentation in strtotime format.")
	RegisterGlobalNS("time", "RFC850", "", clock.RFC850).
		WithType("string").
		WithDescription("time.RFC850 presentation in strtotime format.")
	RegisterGlobalNS("time", "RFC1123", "", clock.RFC1123).
		WithType("string").
		WithDescription("time.RFC1123 presentation in strtotime format.")
	RegisterGlobalNS("time", "RFC1123Z", "", clock.RFC1123Z).
		WithType("string").
		WithDescription("time.RFC1123Z presentation in strtotime format.")
	RegisterGlobalNS("time", "RFC3339", "", clock.RFC3339).
		WithType("string").
		WithDescription("time.RFC3339 presentation in strtotime format.")
	RegisterGlobalNS("time", "RFC3339Nano", "", clock.RFC3339Nano).
		WithType("string").
		WithDescription("time.RFC3339Nano presentation in strtotime format.")
	RegisterGlobalNS("time", "Kitchen", "", clock.Kitchen).
		WithType("string").
		WithDescription("time.Kitchen presentation in strtotime format.")
	RegisterGlobalNS("time", "Stamp", "", clock.Stamp).
		WithType("string").
		WithDescription("time.Stamp presentation in strtotime format.")
	RegisterGlobalNS("time", "StampMilli", "", clock.StampMilli).
		WithType("string").
		WithDescription("time.StampMilli presentation in strtotime format.")
	RegisterGlobalNS("time", "StampMicro", "", clock.StampMicro).
		WithType("string").
		WithDescription("time.StampMicro presentation in strtotime format.")
	RegisterGlobalNS("time", "StampNano", "", clock.StampNano).
		WithType("string").
		WithDescription("time.StampNano presentation in strtotime format.")

	// Register testing stuff.
	RegisterCondOKFnNS("testns", "condHelperOK", func(_ *Ctx, v *any, ok *bool, _ []any) { *v, *ok = 15, true }).
		WithDescription("Testing stuff: don't use in production.")
	RegisterCondOKFnNS("testns", "condHelperNotOK", func(_ *Ctx, v *any, ok *bool, _ []any) { *v, *ok = 17, false }).
		WithDescription("Testing stuff: don't use in production.")
	RegisterCallbackFnNS("testns", "foo", "nop", func(_ *Ctx, _ []any) error { return nil }).
		WithDescription("Testing stuff: don't use in production.")
}
