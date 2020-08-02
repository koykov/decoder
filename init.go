package jsondecoder

import "github.com/koykov/inspector"

func init() {
	RegisterModFn("default", "def", modDefault)

	RegisterGetterFn("crc32", "", getterCrc32)

	RegisterCallbackFn("foo", "", cbFoo)

	inspector.RegisterAssignFn(AssignJsonNode)
}
