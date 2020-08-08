package jsondecoder

import "github.com/koykov/inspector"

func init() {
	RegisterModFn("default", "def", modDefault)

	RegisterGetterFn("crc32", "", getterCrc32)
	RegisterGetterFn("appendTestHistory", "", getterAppendTestHistory)

	RegisterCallbackFn("foo", "", cbFoo)
	RegisterCallbackFn("jsonParseAs", "jsonParse", cbJsonParse)

	inspector.RegisterAssignFn(AssignJsonNode)
}
