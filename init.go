package decoder

import (
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

	// Register testing stuff.
	RegisterCondOKFnNS("testns", "condHelperOK", func(_ *Ctx, v *any, ok *bool, _ []any) { *v, *ok = 15, true }).
		WithDescription("Testing stuff: don't use in production.")
	RegisterCondOKFnNS("testns", "condHelperNotOK", func(_ *Ctx, v *any, ok *bool, _ []any) { *v, *ok = 17, false }).
		WithDescription("Testing stuff: don't use in production.")
	RegisterCallbackFnNS("testns", "foo", "nop", func(_ *Ctx, _ []any) error { return nil }).
		WithDescription("Testing stuff: don't use in production.")
}
