package decoder

import (
	"github.com/koykov/inspector"
)

func init() {
	// Register builtin modifiers.
	RegisterModFn("default", "def", modDefault)
	RegisterModFn("ifThen", "if", modIfThen)
	RegisterModFn("ifThenElse", "ifel", modIfThenElse)
	RegisterModFnNS("bar", "baz", "", func(_ *Ctx, _ *any, _ any, _ []any) error { return nil })

	// Register builtin getter callbacks.
	RegisterGetterFn("crc32", "", getterCrc32)
	RegisterGetterFn("strToInt", "atoi", getterAtoi)
	RegisterGetterFn("strToUint", "atou", getterAtou)
	RegisterGetterFn("strToFloat", "atof", getterAtof)
	RegisterGetterFn("strToBool", "atob", getterAtob)
	RegisterGetterFn("intToStr", "itoa", getterItoa)
	RegisterGetterFn("uintToStr", "utoa", getterUtoa)
	RegisterGetterFn("appendTestHistory", "", getterAppendTestHistory)

	// Register builtin callbacks.
	RegisterCallbackFnNS("testns", "foo", "", func(_ *Ctx, _ []any) error { return nil })

	// Register assign functions.
	inspector.RegisterAssignFn(AssignVectorNode)
}
