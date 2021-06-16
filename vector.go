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
	VectorJson VectorType = iota
	VectorUrl
	VectorXml
	VectorYaml
)

// Assign parser helper to the vector vec according given type.
func ensureHelper(vec vector.Interface, typ VectorType) vector.Interface {
	switch typ {
	case VectorJson:
		vec.SetHelper(&jsonvector.JsonHelper{})
	case VectorUrl:
		vec.SetHelper(&urlvector.URLHelper{})
	case VectorXml:
		// todo set proper helper when xmlvector package will implements.
		vec.SetHelper(nil)
	case VectorYaml:
		// todo set proper helper when yamlvector package will implements.
		vec.SetHelper(nil)
	}
	return vec
}

// Make new vector parser according given type.
func newVector(typ VectorType) vector.Interface {
	switch typ {
	case VectorJson:
		return jsonvector.NewVector()
	case VectorUrl:
		return urlvector.NewVector()
	case VectorXml:
		return xmlvector.NewVector()
	case VectorYaml:
		return yamlvector.NewVector()
	}
	return nil
}
