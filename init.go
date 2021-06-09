package decoder

import "github.com/koykov/inspector"

func init() {
	// Register builtin modifiers.
	RegisterModFn("default", "def", modDefault)
	RegisterModFn("ifThen", "if", modIfThen)
	RegisterModFn("ifThenElse", "ifel", modIfThenElse)

	// Register builtin getter callbacks.
	RegisterGetterFn("crc32", "", getterCrc32)
	RegisterGetterFn("appendTestHistory", "", getterAppendTestHistory)

	// Register builtin callbacks.
	RegisterCallbackFn("foo", "", cbFoo)
	// todo move callback to further bridge package.
	RegisterCallbackFn("jsonParseAs", "jsonParse", cbJsonParse)

	// Register assign functions.
	inspector.RegisterAssignFn(AssignVectorNode)
}
