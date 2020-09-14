package decoder

import "github.com/koykov/inspector"

func init() {
	// Register builtin modifiers.
	RegisterModFn("default", "def", modDefault)

	// Register builtin getter callbacks.
	RegisterGetterFn("crc32", "", getterCrc32)
	RegisterGetterFn("appendTestHistory", "", getterAppendTestHistory)

	// Register builtin callbacks.
	RegisterCallbackFn("foo", "", cbFoo)
	RegisterCallbackFn("jsonParseAs", "jsonParse", cbJsonParse)

	// Register assign functions.
	inspector.RegisterAssignFn(AssignJsonNode)
}
