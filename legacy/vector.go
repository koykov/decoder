package legacy

import (
	"github.com/koykov/decoder"
	"github.com/koykov/jsonvector"
	"github.com/koykov/urlvector"
	"github.com/koykov/vector"
	"github.com/koykov/xmlvector"
	"github.com/koykov/yamlvector"
)

// Assign parser helper to the vector vec according given type.
func ensureHelper(vec vector.Interface, typ decoder.VectorType) vector.Interface {
	switch typ {
	case decoder.VectorJSON:
		vec.SetHelper(jsonvector.Helper{})
	case decoder.VectorURL:
		vec.SetHelper(urlvector.Helper{})
	case decoder.VectorXML:
		vec.SetHelper(xmlvector.Helper{})
	case decoder.VectorYAML:
		// todo set proper helper when yamlvector package will implements.
		vec.SetHelper(nil)
	}
	return vec
}

// Make new vector parser according given type.
func newVector(typ decoder.VectorType) vector.Interface {
	switch typ {
	case decoder.VectorJSON:
		return jsonvector.NewVector()
	case decoder.VectorURL:
		return urlvector.NewVector()
	case decoder.VectorXML:
		return xmlvector.NewVector()
	case decoder.VectorYAML:
		return yamlvector.NewVector()
	}
	return nil
}
