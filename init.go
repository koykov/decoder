package jsondecoder

func init() {
	RegisterModFn("default", "def", modDefault)

	RegisterGetterFn("crc32", "", getterCrc32)
}
