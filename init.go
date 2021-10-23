package decoder

import "github.com/koykov/inspector"

func init() {
	// Register builtin modifiers.
	RegisterModFn("default", "def", modDefault)
	RegisterModFn("ifThen", "if", modIfThen)
	RegisterModFn("ifThenElse", "ifel", modIfThenElse)

	// Register builtin getter callbacks.
	RegisterGetterFn("crc32", "", getterCrc32)
	RegisterGetterFn("strToInt", "atoi", getterAtoi)
	RegisterGetterFn("strToUint", "atou", getterAtou)
	RegisterGetterFn("strToFloat", "atof", getterAtof)
	RegisterGetterFn("strToBool", "atob", getterAtob)
	RegisterGetterFn("appendTestHistory", "", getterAppendTestHistory)

	// Register builtin callbacks.
	RegisterCallbackFn("foo", "", cbFoo)
	RegisterCallbackFn("jsonParseAs", "jsonParse", cbJsonParse)
	RegisterCallbackFn("urlParseAs", "urlParse", cbUrlParse)
	RegisterCallbackFn("xmlParseAs", "xmlParse", cbXmlParse)
	RegisterCallbackFn("yamlParseAs", "yamlParse", cbYamlParse)

	// Register assign functions.
	inspector.RegisterAssignFn(AssignVectorNode)
}
