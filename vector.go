package decoder

import (
	"github.com/koykov/jsonvector"
	"github.com/koykov/urlvector"
	"github.com/koykov/vector"
	"github.com/koykov/xmlvector"
	"github.com/koykov/yamlvector"
)

type VectorType int

const (
	VectorJSON VectorType = iota
	VectorURL
	VectorXML
	VectorYAML

	VectorsSupported = 4
)

// Assign parser helper to the vector vec according given type.
func ensureHelper(vec vector.Interface, typ VectorType) vector.Interface {
	switch typ {
	case VectorJSON:
		vec.SetHelper(jsonvector.Helper{})
	case VectorURL:
		vec.SetHelper(urlvector.Helper{})
	case VectorXML:
		vec.SetHelper(xmlvector.Helper{})
	case VectorYAML:
		// todo set proper helper when yamlvector package will implements.
		vec.SetHelper(nil)
	}
	return vec
}

// Make new vector parser according given type.
func newVector(typ VectorType) vector.Interface {
	switch typ {
	case VectorJSON:
		return jsonvector.NewVector()
	case VectorURL:
		return urlvector.NewVector()
	case VectorXML:
		return xmlvector.NewVector()
	case VectorYAML:
		return yamlvector.NewVector()
	}
	return nil
}
